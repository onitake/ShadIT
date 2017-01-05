export GOPATH=$(shell pwd)

all: bin/server

clean:
	rm -f bin/* pkg/*

bin/server: src/api.go
	go build -o $@ $^
