package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
)

func (s *Service) SendMail(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	var req models.MailerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Erreur JSON:", err)
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		log.Println("❌ Adresse email invalide :", req.Email)
		return httpx.WriteError(w, http.StatusBadRequest, "Invalid email address")
	}
	err = s.mailer.SendTestMail(req.Email)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, err.Error())
	}
	return nil
}
