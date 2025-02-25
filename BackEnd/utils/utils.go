package utils

import (
	"net/http"
)

var mux = http.NewServeMux()

type methodHandlers struct {
	get  func(w http.ResponseWriter, _ *http.Request) error
	post func(w http.ResponseWriter, _ *http.Request) error
}

var handlers = make(map[string]*methodHandlers)

func GET(path string, handlerFunc func(w http.ResponseWriter, _ *http.Request) error) {
	if handlers[path] == nil {
		handlers[path] = &methodHandlers{}
		mux.HandleFunc(path, handleRequest)
	}
	handlers[path].get = handlerFunc
}

func POST(path string, handlerFunc func(w http.ResponseWriter, _ *http.Request) error) {
	if handlers[path] == nil {
		handlers[path] = &methodHandlers{}
		mux.HandleFunc(path, handleRequest)
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
			_ = handler.get(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	case http.MethodPost:
		if handler.post != nil {
			_ = handler.post(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func Start(port string) {
	_ = http.ListenAndServe(port, mux)
}
