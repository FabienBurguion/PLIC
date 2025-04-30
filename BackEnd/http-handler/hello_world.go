package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"log"
	"net/http"
	"time"
)

// GetTime godoc
// @Summary      Get current server time
// @Description  Returns the current server time. If the user is not authenticated, returns a fixed default time.
// @Tags         testing
// @Produce      json
// @Success      200 {string} string "Current time in RFC3339 format"
// @Router       / [get]
func (s *Service) GetTime(w http.ResponseWriter, _ *http.Request, ai models.AuthInfo) error {
	log.Printf("IsConnected: %v\n", ai.IsConnected)
	log.Println("UserId: ", ai.UserID)
	if !ai.IsConnected {
		return httpx.Write(w, http.StatusOK, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	}
	return httpx.Write(w, http.StatusOK, s.clock.Now())
}

// GetHelloWorld godoc
// @Summary      Say Hello
// @Description  Returns a greeting with the provided name
// @Tags         testing
// @Produce      json
// @Param        name query string true "Name to greet"
// @Success      200 {object} models.HelloWorldResponse
// @Router       /hello_world [get]
func (s *Service) GetHelloWorld(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	name := r.URL.Query().Get("name")
	return httpx.Write(w, http.StatusOK, models.HelloWorldResponse{
		Response: "Hello " + name,
	})
}
