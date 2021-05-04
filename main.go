package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Server struct {
	wg *sync.WaitGroup
	db *sql.DB
}

type Redirect struct {
	Key string
	Url string
}

//run in goroutine - connect to db
func (s *Server) connectDb() {
	host, ok := os.LookupEnv("POSTGRES_HOST")

	if !ok {
		log.Fatalf("POSTGRES_HOST not set")
	}

	user, ok := os.LookupEnv("POSTGRES_USER")
	if !ok {
		log.Fatalf("POSTGRES_USER not set")
	}

	password, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		log.Fatalf("POSTGRES_PASSWORD not set")
	}

	pgdb, ok := os.LookupEnv("POSTGRES_DB")
	if !ok {
		log.Fatalf("POSTGRES_DB not set")
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s?sslmode=disable", user, password, host)
	password = ""

	db, err := sql.Open(pgdb, connStr)
	if err != nil {
		log.Fatal(err)
	}
	s.db = db
	s.wg.Done()
}

func NewServer() *Server {
	var wg sync.WaitGroup
	return &Server{
		wg: &wg,
	}
}

//run in goroutine - handle requests
func (s *Server) handleRequests() {
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.PathPrefix("/").
		HandlerFunc(s.getHandler).
		Methods("GET")
	myRouter.PathPrefix("/").
		HandlerFunc(s.postHandler).
		Methods("POST")
	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	log.Fatal(http.ListenAndServe(":8080", myRouter))
	s.wg.Done()
}

func main() {
	server := NewServer()
	fmt.Println("Rest API v2.0 - Mux Routers")
	go server.handleRequests()
	server.wg.Add(1)

	go server.connectDb()
	server.wg.Add(1)
	server.wg.Wait()
}

//callbacks
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the homepage")
	fmt.Println("GET on homepage endpoint")
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	var redirect Redirect
	var query string
	var err error
	var rows *sql.Rows
	path := r.URL.String()

	//check path
	//getAll
	if path == "/" {
		query = "SELECT key, url FROM redirects"
		rows, err = s.db.Query(query)
	} else {
		trimmedPath := strings.TrimPrefix(path, "/")
		query = fmt.Sprintf("SELECT key, url FROM redirects WHERE key = '%s'", trimmedPath)
		err = s.db.QueryRow(query).Scan(&redirect.Key, &redirect.Url)
	}

	fmt.Printf("getHandler request - %+v\n", r)
	fmt.Printf("getHandler path - %v\n", r.URL)

	if err != nil {
		fmt.Fprintf(w, "Error querying postgre, err: %v\n", err)
		log.Printf("Error querying postgre, err: %v\n", err)
		return
	}

	if path == "/" {
		fmt.Fprintf(w, "GET all from db: %+v\n", rows)
		log.Printf("GET all from db: %+v\n", rows)
	} else {
		//redirect user to the redirect.Url
		http.Redirect(w, r, redirect.Url, 308)
		fmt.Fprintf(w, "GET from db: %+v\n", redirect)
		log.Printf("GET from db: %+v\n", redirect)
	}

}

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) {
	var body bytes.Buffer
	var redirect Redirect

	body.ReadFrom(r.Body)
	err := json.Unmarshal(body.Bytes(), &redirect)

	if err != nil {
		fmt.Fprintf(w, "Error unmarshaling: %v\n", body.Bytes())
		log.Printf("Error unmarshaling: %v\n", body.Bytes())
		return
	}

	fmt.Printf("postHandler request - %+v\n", r)
	fmt.Printf("postHandler r.body - %+v\n", body.String())
	fmt.Printf("postHandler redirect - %+v\n", redirect)
	fmt.Printf("postHandler path - %v\n", r.URL)

	//validate
	if redirect.Key == "" || redirect.Url == "" {
		fmt.Fprintf(w, "Received empty value, expected non-empty string, key: %v, value: %v\n",
			redirect.Key, redirect.Url)
		log.Printf("Received empty value, expected non-empty string, key: %v, value: %v\n",
			redirect.Key, redirect.Url)
		return
	}

	query := fmt.Sprintf("INSERT into redirects(key, url) VALUES ('%s', '%s')", redirect.Key, redirect.Url)
	fmt.Printf("Query - %v\n", query)
	first, err := s.db.Exec(query)

	fmt.Printf("%v", first)
	if err != nil {
		fmt.Fprintf(w, "Error querying postgre, err: %v\n", err)
		log.Printf("Error querying postgre, err: %v\n", err)
		return
	}

	fmt.Fprintf(w, "Successfully wrote to DB - key: %v, url: %v\n", redirect.Key, redirect.Url)
	log.Printf("Successfully wrote to DB - key: %v, url: %v\n", redirect.Key, redirect.Url)
}
