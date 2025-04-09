package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"
)

type httpHandler func(http.ResponseWriter, *http.Request, models.AuthInfo) error

type methodHandlers struct {
	get  httpHandler
	post httpHandler
}

var handlers = make(map[string]*methodHandlers)

func (s *Service) GET(path string, handlerFunc httpHandler) {
	if handlers[path] == nil {
		handlers[path] = &methodHandlers{}
		s.server.HandleFunc(path, handleRequest)
	}
	handlers[path].get = handlerFunc
}

func (s *Service) POST(path string, handlerFunc httpHandler) {
	if handlers[path] == nil {
		handlers[path] = &methodHandlers{}
		s.server.HandleFunc(path, handleRequest)
	}
	handlers[path].post = handlerFunc
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	handler := handlers[r.URL.Path]
	if handler == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if handler.get != nil {
			_ = handler.get(w, r, models.AuthInfo{})
		} else {
			_ = httpx.WriteError(w, http.StatusMethodNotAllowed, httpx.MethodNotAllowedError)
		}
	case http.MethodPost:
		if handler.post != nil {
			_ = handler.post(w, r, models.AuthInfo{})
		} else {
			_ = httpx.WriteError(w, http.StatusMethodNotAllowed, httpx.MethodNotAllowedError)
		}
	default:
		_ = httpx.WriteError(w, http.StatusMethodNotAllowed, httpx.MethodNotAllowedError)
	}
}
