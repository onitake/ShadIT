export GOPATH=$(shell pwd)

all: bin/server

clean:
	rm -f bin/* pkg/*

bin/server: src/api.go src/main.go src/model.go src/gpio.go
	go build -o $@ $^
