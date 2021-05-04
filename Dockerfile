FROM golang

WORKDIR /go/src/app
COPY . .
RUN go mod tidy
RUN go build -o go-redirector

ENTRYPOINT ["./go-redirector"]
