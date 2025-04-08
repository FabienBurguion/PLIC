package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
)

var jwtSecret = os.Getenv("JWT_SECRET")

func (s *Service) withAuthentication(handler func(http.ResponseWriter, *http.Request, models.AuthInfo) error) func(w http.ResponseWriter, r *http.Request, info models.AuthInfo) error {
	return func(w http.ResponseWriter, r *http.Request, info models.AuthInfo) error {
		auth := models.AuthInfo{IsConnected: false}

		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return jwtSecret, nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if userID, ok := claims["user_id"].(string); ok {
						auth.IsConnected = true
						auth.UserID = userID
					}
				}
			}
		}

		if err := handler(w, r, auth); err != nil {
			return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
		}
		return nil
	}
}
