package main

import (
	"log"
	"net/url"
	"net/http"
	"encoding/json"
)

var (
	ErrInternal = []byte("{'error':'internal'}")
	ErrNotImplemented = []byte("{'error':'not_implemented'}")
	ErrInvalidObject = []byte("{'error':'invalid_object'}")
)

type Endpoint interface {
	Handle(path []string, query url.Values) ([]byte, int)
}

type RootEndpoint struct {
	State *ShadeState
}

func NewRootEndpoint(state *ShadeState) Endpoint {
	return &RootEndpoint{
		State: state,
	}
}

func (ep *RootEndpoint) Handle(path []string, query url.Values) ([]byte, int) {
	//log.Printf("len(path)=%d path[0]=%s path[1]=%s\n", len(path), path[0], path[1])
	if (len(path) >= 1) {
		if (path[0] == "") {
			shades := make(map[string]map[string]interface{})
			for _, shade := range(ep.State.Shades) {
				shades[shade.Name] = map[string]interface{}{
					"name": shade.Name,
					"position": shade.Position,
					"angle": shade.Angle,
				}
			}
			response, err := json.Marshal(map[string]interface{}{
				"shades": shades,
				"error": "none",
			})
			if (err == nil) {
				return response, http.StatusOK
			} else {
				log.Print(err)
				return ErrInternal, http.StatusInternalServerError
			}
		} else {
			shade := ep.State.Shades[path[0]]
			if (shade != nil) {
				if (len(path) > 1) {
					return ErrNotImplemented, http.StatusNotFound
				} else {
					response, err := json.Marshal(map[string]interface{}{
						"name": shade.Name,
						"position": shade.Position,
						"angle": shade.Angle,
					})
					if (err == nil) {
						return response, http.StatusOK
					} else {
						log.Print(err)
						return ErrInternal, http.StatusInternalServerError
					}
				}
			} else {
				log.Printf("restreamer: object %s not found\n", path[0])
				return ErrInvalidObject, http.StatusNotFound
			}
		}
	} else {
		log.Printf("restreamer: empty argument list\n", path[0])
		return ErrInvalidObject, http.StatusNotFound
	}
}
