package main

import (
	"log"
	"os"
	"strings"
	"net/http"
	"encoding/json"
)

type Configuration struct {
	Listen string
	UpTime int
	DownTime int
	FlipTime int
	Shades []struct {
		Name string
		GpioUp int
		GpioDown int
	}
}

type ShadeServer struct {
	Root Endpoint
}

func NewShadeServer(state *ShadeState) *ShadeServer {
	return &ShadeServer{
		Root: NewRootEndpoint(state),
	}
}

func (server *ShadeServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	path := strings.Split(request.URL.Path, "/")
	log.Printf("len(path)=%d path[0]=%s path[1]=%s\n", len(path), path[0], path[1])
	server.Root.Handle(writer, request)
}

func main() {
	var configname string
	if len(os.Args) > 1 {
		configname = os.Args[1]
	} else {
		configname = "server.json"
	}
	
	configfile, err := os.Open(configname)
	if err != nil {
		log.Fatal("Can't read configuration from server.json: ", err)
	}
	decoder := json.NewDecoder(configfile)
	config := Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error parsing configuration: ", err)
	}
	configfile.Close()

	state := NewShadeState(&config)
	server := NewShadeServer(state)
	log.Fatal(http.ListenAndServe(config.Listen, server))
}
