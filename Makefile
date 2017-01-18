export GOPATH=$(shell pwd)

all: bin/shudder

clean:
	rm -f bin/* pkg/*

bin/shudder: src/api.go src/main.go src/model.go src/gpio.go
	go build -o $@ $^
