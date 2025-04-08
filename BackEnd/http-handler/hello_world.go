package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"
)

func (s *Service) GetTime(w http.ResponseWriter, _ *http.Request, _ models.AuthInfo) error {
	return httpx.Write(w, http.StatusOK, s.clock.Now())
}

func (s *Service) GetHelloWorld(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	name := r.URL.Query().Get("name")
	return httpx.Write(w, http.StatusOK, models.HelloWorldResponse{
		Response: "Hello " + name,
	})
}
