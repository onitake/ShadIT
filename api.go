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

func jsonResponse(data interface{}, code int) ([]byte, int) {
	response, err := json.Marshal(data)
	if err == nil {
		return response, code
	} else {
		log.Print(err)
		response, err = json.Marshal(map[string]string{
			"error": ErrInternal,
		})
		if (err == nil) {
			return response, http.StatusInternalServerError
		} else {
			log.Print(err)
			return []byte{}, http.StatusInternalServerError
		}
	}
}

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
		return jsonResponse(map[string]interface{}{
			"children": ep.Children(),
		}, http.StatusOK)
	} else {
		child := ep.children[path[0]]
		if child != nil {
			return child.Handle(path[1:], query)
		} else {
			log.Printf("restreamer: unknown child %s\n", path[0])
			return jsonResponse(map[string]interface{}{
				"error": ErrInvalidObject,
			}, http.StatusNotFound)
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
		return jsonResponse(map[string]interface{}{
			"name": shutter.Name,
			"children": ep.Children(),
			"position": shutter.Position,
			"angle": shutter.Angle,
		}, http.StatusOK)
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
			return jsonResponse(map[string]interface{}{
				"name": shutter.Name,
				"angle": shutter.Angle,
			}, http.StatusOK)
		} else {
			log.Print(err)
			return jsonResponse(map[string]interface{}{
				"error": ErrInvalidArgument,
				"args": []interface{}{
					map[string]interface{}{
						"name": "angle",
						"type": "float",
						"range_from": 0.0,
						"range_to": 1.0,
					},
				},
			}, http.StatusBadRequest)
		}
	} else {
		log.Printf("restreamer: unknown child %s\n", path[0])
		return jsonResponse(map[string]interface{}{
			"error": ErrInvalidObject,
		}, http.StatusNotFound)
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
			return jsonResponse(map[string]interface{}{
				"name": shutter.Name,
				"position": shutter.Position,
			}, http.StatusOK)
		} else {
			log.Print(err)
			return jsonResponse(map[string]interface{}{
				"error": ErrInvalidArgument,
				"args": []interface{}{
					map[string]interface{}{
						"name": "position",
						"type": "float",
						"range_from": 0.0,
						"range_to": 100.0,
					},
				},
			}, http.StatusBadRequest)
		}
	} else {
		log.Printf("restreamer: unknown child %s\n", path[0])
		return jsonResponse(map[string]interface{}{
			"error": ErrInvalidObject,
		}, http.StatusNotFound)
	}
}
