package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"
)

type httpHandler func(http.ResponseWriter, *http.Request, models.AuthInfo) error

func (s *Service) GET(path string, handlerFunc httpHandler) {
	s.server.Get(path, func(w http.ResponseWriter, r *http.Request) {
		if err := handlerFunc(w, r, models.AuthInfo{}); err != nil {
			_ = httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		}
	})
}

func (s *Service) POST(path string, handlerFunc httpHandler) {
	s.server.Post(path, func(w http.ResponseWriter, r *http.Request) {
		if err := handlerFunc(w, r, models.AuthInfo{}); err != nil {
			_ = httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		}
	})
}
