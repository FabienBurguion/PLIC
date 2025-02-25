package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"
)

func GetHelloWorld(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get("name")
	return httpx.Write(w, http.StatusOK, models.HelloWorldResponse{
		Response: "Hello " + name,
	})
}
