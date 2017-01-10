package main

import (
	"log"
	"strconv"
	"net/url"
	"net/http"
	"encoding/json"
)

var (
	ErrInternal = "internal"
	ErrNotImplemented = "not_implemented"
	ErrInvalidObject = "invalid_object"
	ErrInvalidArgument = "invalid_argument"
)

type Endpoint interface {
	Handle(path []string, query url.Values) ([]byte, int)
}

type TreeEndpoint struct {
	children map[string]Endpoint
}

func NewTreeEndpoint() *TreeEndpoint {
	return &TreeEndpoint{
		children: make(map[string]Endpoint),
	}
}

func (ep *TreeEndpoint) Children() []string {
	keys := make([]string, len(ep.children))
	i := 0
	for key := range ep.children {
		keys[i] = key
		i++
	}
	return keys
}

func (ep *TreeEndpoint) Handle(path []string, query url.Values) ([]byte, int) {
	if path == nil || len(path) == 0 || path[0] == "" {
		response, err := json.Marshal(map[string]interface{}{
			"children": ep.Children(),
			"error": "none",
		})
		if err == nil {
			return response, http.StatusOK
		} else {
			log.Print(err)
			return []byte("{\"error\":\"" + ErrInternal + "\"}"), http.StatusInternalServerError
		}
	} else {
		child := ep.children[path[0]]
		if child != nil {
			return child.Handle(path[1:], query)
		} else {
			log.Printf("restreamer: unknown child %s\n", path[0])
			return []byte("{\"error\":\"" + ErrInvalidObject + "\"}"), http.StatusNotFound
		}
	}
}

type RootEndpoint struct {
	*TreeEndpoint
	state *ShutterState
}

func NewRootEndpoint(state *ShutterState) *RootEndpoint {
	ep := &RootEndpoint{
		TreeEndpoint: NewTreeEndpoint(),
		state: state,
	}
	for key := range state.Shutters {
		ep.children[key] = NewShutterEndpoint(state, key)
	}
	return ep
}

type ShutterEndpoint struct {
	*TreeEndpoint
	state *ShutterState
	name string
}

func NewShutterEndpoint(state *ShutterState, name string) *ShutterEndpoint {
	ep := &ShutterEndpoint{
		TreeEndpoint: NewTreeEndpoint(),
		state: state,
		name: name,
	}
	ep.children["flip"] = NewFlipEndpoint(state, name)
	ep.children["move"] = NewMoveEndpoint(state, name)
	return ep
}

func (ep *ShutterEndpoint) Handle(path []string, query url.Values) ([]byte, int) {
	//log.Printf("len(path)=%d path[0]=%s path[1]=%s\n", len(path), path[0], path[1])
	if path == nil || len(path) == 0 || path[0] == "" {
		shutter := ep.state.Shutters[ep.name]
		response, err := json.Marshal(map[string]interface{}{
			"name": shutter.Name,
			"children": ep.Children(),
			"position": shutter.Position,
			"angle": shutter.Angle,
		})
		if err == nil {
			return response, http.StatusOK
		} else {
			log.Print(err)
			return []byte("{\"error\":\"" + ErrInternal + "\"}"), http.StatusInternalServerError
		}
	} else {
		return ep.TreeEndpoint.Handle(path, query)
	}
}

func (ep *ShutterEndpoint) Children() []string {
	return make([]string, 0)
}

func (ep *ShutterEndpoint) Child(key string) Endpoint {
	return nil
}

type FlipEndpoint struct {
	state *ShutterState
	name string
}

func NewFlipEndpoint(state *ShutterState, name string) *FlipEndpoint {
	return &FlipEndpoint{
		state: state,
		name: name,
	}
}

func (ep *FlipEndpoint) Handle(path []string, query url.Values) ([]byte, int) {
	//log.Printf("len(path)=%d path[0]=%s path[1]=%s\n", len(path), path[0], path[1])
	if path == nil || len(path) == 0 || path[0] == "" {
		shutter := ep.state.Shutters[ep.name]
		angle, err := strconv.ParseFloat(query.Get("angle"), 32)
		if err == nil {
			// TODO use a queue instead of just running this synchronously
			shutter.Flip(float32(angle))
			response, err := json.Marshal(map[string]interface{}{
				"name": shutter.Name,
				"angle": shutter.Angle,
			})
			if err == nil {
				return response, http.StatusOK
			} else {
				log.Print(err)
				return []byte("{'error':'" + ErrInternal + "'}"), http.StatusInternalServerError
			}
		} else {
			log.Print(err)
			return []byte("{\"error\":\"" + ErrInvalidArgument + "\",\"args\":[{\"name\":\"angle\",\"type\":\"float\",\"range\":\"0..1\"}]}"), http.StatusBadRequest
		}
	} else {
		log.Printf("restreamer: unknown child %s\n", path[0])
		return []byte("{\"error\":\"" + ErrInvalidObject + "\"}"), http.StatusNotFound
	}
}

type MoveEndpoint struct {
	state *ShutterState
	name string
}

func NewMoveEndpoint(state *ShutterState, name string) *MoveEndpoint {
	return &MoveEndpoint{
		state: state,
		name: name,
	}
}

func (ep *MoveEndpoint) Handle(path []string, query url.Values) ([]byte, int) {
	//log.Printf("len(path)=%d path[0]=%s path[1]=%s\n", len(path), path[0], path[1])
	if path == nil || len(path) == 0 || path[0] == "" {
		shutter := ep.state.Shutters[ep.name]
		position, err := strconv.ParseFloat(query.Get("position"), 32)
		if err == nil {
			// TODO use a queue instead of just running this synchronously
			shutter.Move(float32(position))
			response, err := json.Marshal(map[string]interface{}{
				"name": shutter.Name,
				"position": shutter.Position,
			})
			if err == nil {
				return response, http.StatusOK
			} else {
				log.Print(err)
				return []byte("{\"error\":\"" + ErrInternal + "\"}"), http.StatusInternalServerError
			}
		} else {
			log.Print(err)
			return []byte("{\"error\":\"" + ErrInvalidArgument + "\",\"args\":[{\"name\":\"position\",\"type\":\"float\",\"range\":\"0..100\"}]}"), http.StatusBadRequest
		}
	} else {
		log.Printf("restreamer: unknown child %s\n", path[0])
		return []byte("{\"error\":\"" + ErrInvalidObject + "\"}"), http.StatusNotFound
	}
}
