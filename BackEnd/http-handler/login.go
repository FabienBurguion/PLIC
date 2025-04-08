package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
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
	ctx := r.Context()
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	user, err := s.db.GetUserWithUsername(ctx, req.Username)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	if user != nil {
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	newUser := models.DBUsers{
		Id:       uuid.NewString(),
		Username: req.Username,
		Password: string(hashedPassword),
	}
	err = s.db.CreateUser(ctx, newUser)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	token, err := GenerateJWT(newUser.Id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	return httpx.Write(w, http.StatusCreated, models.LoginResponse{Token: token})
}
