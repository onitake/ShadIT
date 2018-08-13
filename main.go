/* Copyright (c) 2017-2018, Gregor Riepl <onitake@gmail.com>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 *     Redistributions of source code must retain the above copyright notice,
 *     this list of conditions and the following disclaimer.
 *
 *     Redistributions in binary form must reproduce the above copyright notice,
 *     this list of conditions and the following disclaimer in the documentation
 *     and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
 * ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
 * ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

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
