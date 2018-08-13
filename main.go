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
	Shutters []struct {
		Name string
		GpioUp string
		GpioDown string
	}
}

type ShutterServer struct {
	Root Endpoint
}

func NewShutterServer(state *ShutterState) *ShutterServer {
	return &ShutterServer{
		Root: NewRootEndpoint(state),
	}
}

func (server *ShutterServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	path := strings.Split(request.URL.Path, "/")
	// strip the empty string before the path separator
	if (len(path) >= 1 && path[0] == "") {
		path = path[1:]
	}
	//log.Printf("len(path)=%d path[0]=%s path[1]=%s\n", len(path), path[0], path[1])
	
	response, status := server.Root.Handle(path, request.URL.Query())
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(status);
	writer.Write(response)
}

func main() {
	var configname string
	if len(os.Args) > 1 {
		configname = os.Args[1]
	} else {
		configname = "config.json"
	}
	
	configfile, err := os.Open(configname)
	if err != nil {
		log.Fatal("Can't read configuration from config.json: ", err)
	}
	decoder := json.NewDecoder(configfile)
	config := Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error parsing configuration: ", err)
	}
	configfile.Close()

	state, err := NewShutterState(&config)
	if err != nil {
		log.Fatal("Error creating state object: ", err)
	}
	server := NewShutterServer(state)
	log.Fatal(http.ListenAndServe(config.Listen, server))
}
