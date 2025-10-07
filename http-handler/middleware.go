package main

import (
	"PLIC/models"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

var (
	visitors  = make(map[string]*rate.Limiter)
	mu        sync.Mutex
	jwtSecret = os.Getenv("JWT_SECRET")
)

func getRealIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xrip := r.Header.Get("X-Real-Ip"); xrip != "" {
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
		limiter = rate.NewLimiter(5, 10)
		visitors[ip] = limiter
	}
	return limiter
}

func withRateLimit(handler httpHandler) httpHandler {
	return func(w http.ResponseWriter, r *http.Request, info models.AuthInfo) error {
		ip := getRealIP(r)
		limiter := getVisitor(ip)

		if !limiter.Allow() {
			log.Warn().
				Str("ip", ip).
				Str("path", r.URL.Path).
				Msg("rate limit exceeded")
			http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
			return nil
		}

		return handler(w, r, info)
	}
}

func withAuthentication(handler httpHandler) httpHandler {
	return withRateLimit(func(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
		ip := getRealIP(r)
		logger := log.With().
			Str("middleware", "auth").
			Str("ip", ip).
			Str("path", r.URL.Path).
			Logger()

		auth := models.AuthInfo{IsConnected: false}
		authHeader := r.Header.Get("Authorization")

		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					logger.Warn().Str("alg", fmt.Sprint(token.Header["alg"])).Msg("unexpected signing method")
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				logger.Warn().Err(err).Msg("invalid JWT token")
			} else if token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if userID, ok := claims["user_id"].(string); ok {
						auth.IsConnected = true
						auth.UserID = userID
					}
				}
			}
		}

		if auth.IsConnected {
			logger.Info().Str("user_id", auth.UserID).Msg("authenticated request")
		} else {
			logger.Info().Msg("unauthenticated request")
		}

		return handler(w, r, auth)
	})
}
