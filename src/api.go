package main

import (
	"log"
	"strings"
	"net/http"
	"encoding/json"
)

type Endpoint interface {
	Handle(writer http.ResponseWriter, request *http.Request)
}

type RootEndpoint struct {
	State *ShadeState
}

func NewRootEndpoint(state *ShadeState) Endpoint {
	return &RootEndpoint{
		State: state,
	}
}

func (ep *RootEndpoint) Handle(writer http.ResponseWriter, request *http.Request) {
	path := strings.Split(request.URL.Path, "/")
	log.Printf("len(path)=%d path[0]=%s path[1]=%s\n", len(path), path[0], path[1])
	if (len(path) == 0) {
		shades := make(map[string]map[string]interface{})
		for _, shade := range(ep.State.Shades) {
			shades[shade.Name] = map[string]interface{}{
				"name": shade.Name,
				"position": shade.Position,
				"angle": shade.Angle,
			}
		}
		writer.Header().Add("Content-Type", "application/json")
		response, error := json.Marshal(map[string]interface{}{
			"shades": shades,
			"error": "none",
		})
		if (error == nil) {
			writer.WriteHeader(http.StatusOK);
			writer.Write(response)
		} else {
			writer.WriteHeader(http.StatusInternalServerError);
			writer.Write([]byte("{'error':'internal'}"))
			log.Print(error)
		}
	} else {
		shade := ep.State.Shades[path[0]]
		if (shade != nil) {
			if (len(path) > 1) {
				writer.Header().Add("Content-Type", "application/json")
				writer.WriteHeader(http.StatusNotFound);
				writer.Write([]byte("{'error':'not_implemented'}"))
			} else {
				writer.Header().Add("Content-Type", "application/json")
				response, error := json.Marshal(map[string]interface{}{
					"name": shade.Name,
					"position": shade.Position,
					"angle": shade.Angle,
				})
				if (error == nil) {
					writer.WriteHeader(http.StatusOK);
					writer.Write(response)
				} else {
					writer.WriteHeader(http.StatusInternalServerError);
					writer.Write([]byte("{'error':'internal'}"))
					log.Print(error)
				}
			}
		} else {
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(http.StatusNotFound);
			writer.Write([]byte("{'error':'invalid_object'}"))
		}
	}
}
