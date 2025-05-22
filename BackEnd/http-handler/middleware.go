package main

import (
	"PLIC/models"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

var jwtSecret = os.Getenv("JWT_SECRET")

func getRealIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	xrip := r.Header.Get("X-Real-Ip")
	if xrip != "" {
		return strings.TrimSpace(xrip)
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(1, 3)
		visitors[ip] = limiter
	}
	return limiter
}

func withRateLimit(handler httpHandler) httpHandler {
	return func(w http.ResponseWriter, r *http.Request, info models.AuthInfo) error {
		ip := getRealIP(r)

		limiter := getVisitor(ip)
		if !limiter.Allow() {
			http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
			return nil
		}

		return handler(w, r, info)
	}
}

func withAuthentication(handler httpHandler) httpHandler {
	return withRateLimit(func(w http.ResponseWriter, r *http.Request, info models.AuthInfo) error {
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
	})
}
