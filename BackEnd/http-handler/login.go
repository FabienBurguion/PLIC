package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

var tokenDuration = time.Now().Add(24 * time.Hour).Unix()

func GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     tokenDuration,
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtSecret))
}

func (s *Service) Login(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	ctx := r.Context()
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	user, err := s.db.GetUserWithUsername(ctx, req.Username)
	if err != nil || user == nil {
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}
	token, err := GenerateJWT(user.Id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	return httpx.Write(w, http.StatusOK, models.LoginResponse{Token: token})
}

func (s *Service) Register(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	log.Println("Entering Register")
	ctx := r.Context()
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Erreur JSON:", err)
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	user, err := s.db.GetUserWithUsername(ctx, req.Username)
	if err != nil {
		log.Println("Erreur DB:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	if user != nil {
		log.Println("Erreur DB: User exists")
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Erreur hash:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	newUser := models.DBUsers{
		Id:       uuid.NewString(),
		Username: req.Username,
		Password: string(hashedPassword),
	}
	token, err := GenerateJWT(newUser.Id)
	if err != nil {
		log.Println("Erreur Token:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	err = s.db.CreateUser(ctx, newUser)
	if err != nil {
		log.Println("Erreur DB à la création:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	return httpx.Write(w, http.StatusCreated, models.LoginResponse{Token: token})
}
