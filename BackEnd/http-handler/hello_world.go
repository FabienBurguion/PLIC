package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"
)

func (s *Service) GetHelloWorldBasic(w http.ResponseWriter, _ *http.Request) error {
	return httpx.Write(w, http.StatusOK, s.clock.Now())
}

func (s *Service) GetHelloWorld(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get("name")
	return httpx.Write(w, http.StatusOK, models.HelloWorldResponse{
		Response: "Hello " + name,
	})
}

func (s *Service) CreateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	id := r.URL.Query().Get("id")
	err := s.db.CreateUser(ctx, models.DBUser{
		Id:    id,
		Name:  "A name",
		Email: "An email",
	})
	if err != nil {
		return err
	}
	return httpx.Write(w, http.StatusCreated, nil)
}
