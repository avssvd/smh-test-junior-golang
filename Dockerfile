FROM golang:1.17

WORKDIR /go/src/app
COPY *.go .env go.mod go.sum ./

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["/bin/sh", "-c", "go run ."]