package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// GetTime godoc
// @Summary      Get current server time
// @Description  Returns the current server time. If the user is not authenticated, returns a fixed default time.
// @Tags         testing
// @Produce      json
// @Success      200 {string} string "Current time in RFC3339 format"
// @Router       / [get]
func (s *Service) GetTime(w http.ResponseWriter, _ *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "GetTime").
		Str("user_id", ai.UserID).
		Bool("is_connected", ai.IsConnected).
		Logger()

	baseLogger.Info().Msg("entering GetTime")

	if !ai.IsConnected {
		baseLogger.Info().Msg("unauthenticated request — returning default fixed time")
		return httpx.Write(w, http.StatusOK, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	}

	currentTime := s.clock.Now()
	baseLogger.Info().Time("current_time", currentTime).Msg("authenticated request — returning current time")
	return httpx.Write(w, http.StatusOK, currentTime)
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

	logger := log.With().
		Str("method", "GetHelloWorld").
		Str("name", name).
		Logger()

	logger.Info().Msg("entering GetHelloWorld")

	response := models.HelloWorldResponse{
		Response: "Hello " + name,
	}

	logger.Info().Str("response", response.Response).Msg("hello world generated successfully")
	return httpx.Write(w, http.StatusOK, response)
}
