# go-redirector

## About
mnemonic redirects, i.e `localhost/yt` `youtube.com` and can be shortened through an alias such as `go` and therefore becomes `go/yt`.

Currently the web API only supports `GET` and `POST` operations, and there is no front end.

## How to run
```sh
docker-compose up
```

There's a docker-compose.yml file that can spin up both the server and the postgreSQL database.

## Examples
Add a redirect
```sh
curl -d '{"key":"yt","url":"youtube.com"}' -H "Content-Type: application/json" localhost
```
