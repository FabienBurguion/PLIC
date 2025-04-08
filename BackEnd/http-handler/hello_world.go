package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"log"
	"net/http"
	"time"
)

func (s *Service) GetTime(w http.ResponseWriter, _ *http.Request, ai models.AuthInfo) error {
	log.Printf("IsConnected: %v\n", ai.IsConnected)
	log.Println("UserId: ", ai.UserID)
	if !ai.IsConnected {
		return httpx.Write(w, http.StatusOK, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	}
	return httpx.Write(w, http.StatusOK, s.clock.Now())
}

func (s *Service) GetHelloWorld(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	name := r.URL.Query().Get("name")
	return httpx.Write(w, http.StatusOK, models.HelloWorldResponse{
		Response: "Hello " + name,
	})
}
