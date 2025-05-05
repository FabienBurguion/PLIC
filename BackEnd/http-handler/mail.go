package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"crypto/rand"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/big"
	"net/http"
	"net/mail"
)

const passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generatePassword(length int) (string, error) {
	password := make([]byte, length)
	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordCharset))))
		if err != nil {
			return "", err
		}
		password[i] = passwordCharset[num.Int64()]
	}
	return string(password), nil
}

// ForgetPassword godoc
// @Summary      Request password reset
// @Description  Generate a new password and send it via email to the user if the account exists
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.MailerRequest true "Email of the user"
// @Success      200 {object} nil "Success even if user does not exist (for security)"
// @Failure      400 {object} models.Error "Bad request (invalid JSON or email format)"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /forget-password [post]
func (s *Service) ForgetPassword(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	var req models.MailerRequest
	ctx := r.Context()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Erreur JSON:", err)
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		log.Println("❌ Adresse email invalide :", req.Email)
		return httpx.WriteError(w, http.StatusBadRequest, "Invalid email address")
	}
	user, err := s.db.GetUserByUsername(ctx, req.Email)
	if err != nil {
		log.Println("Erreur DB:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	if user == nil {
		log.Println("User does not exist")
		return httpx.Write(w, http.StatusOK, nil)
	}
	newPassword, err := generatePassword(12)
	if err != nil {
		log.Println("Erreur génération mot de passe :", err)
		return httpx.WriteError(w, http.StatusInternalServerError, "Could not generate password")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Erreur hash:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	err = s.db.ChangePassword(ctx, req.Email, string(passwordHash))
	if err != nil {
		log.Println("Erreur DB au changement de password:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	err = s.mailer.SendPasswordForgotMail(req.Email, newPassword)
	if err != nil {
		log.Println("Erreur envoi mail:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	return httpx.Write(w, http.StatusOK, nil)
}
