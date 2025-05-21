package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"crypto/rand"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/big"
	"net/http"
	"net/mail"
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

// Login godoc
// @Summary      Login a user
// @Description  Authenticate a user with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "User credentials"
// @Success      200 {object} models.LoginResponse
// @Failure      400 {object} models.Error "Bad request"
// @Failure      401 {object} models.Error "Invalid credentials"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /login [post]
func (s *Service) Login(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	ctx := r.Context()
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	user, err := s.db.GetUserByUsername(ctx, req.Username)
	if err != nil {
		log.Println("Error getting the user")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	if user == nil {
		log.Println("No user found")
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Println("Error comparing password")
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}
	token, err := GenerateJWT(user.Id)
	if err != nil {
		log.Println("Error generating token")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	return httpx.Write(w, http.StatusOK, models.LoginResponse{Token: token})
}

// Register godoc
// @Summary      Register a new user
// @Description  Register a user with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.RegisterRequest true "User credentials"
// @Success      201 {object} models.LoginResponse
// @Failure      400 {object} models.Error "Bad request"
// @Failure      401 {object} models.Error "User already exists"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /register [post]
func (s *Service) Register(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	log.Println("Entering Register")
	ctx := r.Context()
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Erreur JSON:", err)
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	if req.Email == "" || req.Password == "" {
		log.Printf("Username or Password empty")
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	user, err := s.db.GetUserByEmail(ctx, req.Email)
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
	id := uuid.NewString()
	newUser := models.DBUsers{
		Id:        id,
		Username:  "user" + id,
		Email:     req.Email,
		Bio:       nil,
		Password:  string(hashedPassword),
		CreatedAt: s.clock.Now(),
		UpdatedAt: s.clock.Now(),
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

// ChangePassword godoc
// @Summary      Change password for authenticated user
// @Description  Allows a connected user to change their password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.ChangePasswordRequest true "New password payload"
// @Success      200 {object} nil "Password changed successfully"
// @Failure      400 {object} models.Error "Bad request (invalid JSON)"
// @Failure      401 {object} models.Error "Unauthorized (not connected or user not found)"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /change-password [post]
func (s *Service) ChangePassword(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}
	ctx := r.Context()
	user, err := s.db.GetUserById(ctx, ai.UserID)
	if err != nil {
		log.Println("Erreur DB:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	if user == nil {
		log.Println("User does not exist")
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}
	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Erreur JSON:", err)
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Erreur hash:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	err = s.db.ChangePassword(ctx, user.Username, string(passwordHash))
	if err != nil {
		log.Println("Erreur DB au changement de password:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	return httpx.Write(w, http.StatusOK, nil)
}
