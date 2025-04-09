package main

import (
	"PLIC/models"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"strings"
)

var jwtSecret = os.Getenv("JWT_SECRET")

func withAuthentication(handler httpHandler) httpHandler {
	return func(w http.ResponseWriter, r *http.Request, info models.AuthInfo) error {
		log.Println("Entering authent middleware")
		auth := models.AuthInfo{IsConnected: false}

		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
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
		return handler(w, r, auth)
	}
}
