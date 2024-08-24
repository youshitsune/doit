VERSION 0.8
FROM golang:alpine
WORKDIR /doit

build:
    COPY go.mod go.sum .
    COPY main.go .
    RUN go build -o output/doit
    SAVE ARTIFACT output/doit AS LOCAL doit

podman:
    COPY +build/doit .
    ENTRYPOINT ["/doit/doit"]
    SAVE IMAGE doit:latest
