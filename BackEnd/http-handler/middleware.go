package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
)

var jwtSecret = os.Getenv("JWT_SECRET")

func (s *Service) withAuthentification(httpHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authHeader := r.Header.Get("Authorization")
		if (authHeader == "") || (authHeader == "Bearer") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		res, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error")
			}
			return jwtSecret, nil
		})
		if err != nil || !res.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		claims, ok := res.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userId, ok := claims["userId"].(string)
		userExist, err := s.db.CheckUserExist(ctx, userId)
		if !ok || !userExist || err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	})
}
