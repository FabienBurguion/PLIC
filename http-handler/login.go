package main

import (
	"PLIC/database"
	"PLIC/httpx"
	"PLIC/models"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/mail"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
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
	baseLogger := log.With().Logger()

	ctx := r.Context()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		baseLogger.Error().Err(err).Msg("invalid JSON body")
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}

	logger := baseLogger.With().
		Str("username", req.Username).
		Bool("has_password", req.Password != "").
		Logger()

	logger.Info().Msg("entering Login")

	user, err := s.db.GetUserByUsername(ctx, req.Username)
	if err != nil {
		logger.Error().Err(err).Msg("db get user by username failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	if user == nil {
		logger.Warn().Msg("user not found")
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Warn().Err(err).Msg("password comparison failed")
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}

	token, err := GenerateJWT(user.Id)
	if err != nil {
		logger.Error().Err(err).Msg("JWT generation failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	logger.Info().Str("user_id", user.Id).Msg("login succeeded")
	return httpx.Write(w, http.StatusOK, models.LoginResponse{Token: token})
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
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
func (s *Service) Register(w http.ResponseWriter, r *http.Request) error {
	baseLogger := log.With().Logger()

	ctx := r.Context()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		baseLogger.Error().Err(err).Msg("invalid JSON body")
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}

	logger := baseLogger.With().
		Str("email", req.Email).
		Str("username", req.Username).
		Bool("has_bio", req.Bio != nil).
		Logger()

	logger.Info().Msg("entering Register")

	if !isValidEmail(req.Email) {
		logger.Warn().Msg("invalid email")
		return httpx.WriteError(w, http.StatusBadRequest, "Invalid Email")
	}
	if req.Password == "" || req.Username == "" {
		logger.Warn().Msg("missing username or password")
		return httpx.WriteError(w, http.StatusBadRequest, "Password and username shouldn't be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			logger.Warn().Err(err).Msg("password too long")
			return httpx.WriteError(w, http.StatusBadRequest, "The password is too long")
		}
		logger.Error().Err(err).Msg("password hashing failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	newUser := models.DBUsers{
		Id:        uuid.NewString(),
		Username:  req.Username,
		Email:     req.Email,
		Bio:       req.Bio,
		Password:  string(hashedPassword),
		CreatedAt: s.clock.Now(),
		UpdatedAt: s.clock.Now(),
	}

	if err := s.db.CreateUser(ctx, newUser); err != nil {
		switch {
		case errors.Is(err, database.ErrEmailTaken):
			logger.Warn().Err(err).Msg("email already taken")
			return httpx.WriteError(w, http.StatusConflict, "email_taken")
		case errors.Is(err, database.ErrUsernameTaken):
			logger.Warn().Err(err).Msg("username already taken")
			return httpx.WriteError(w, http.StatusConflict, "username_taken")
		default:
			logger.Error().Err(err).Msg("db create user failed")
			return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
		}
	}

	token, err := GenerateJWT(newUser.Id)
	if err != nil {
		logger.Error().Err(err).Msg("JWT generation failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	logger.Info().Str("user_id", newUser.Id).Msg("user registered successfully")
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

func generateResetToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func parseResetToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email, ok := claims["email"].(string)
		if !ok {
			return "", jwt.ErrTokenMalformed
		}
		return email, nil
	}

	return "", fmt.Errorf("invalid token claims: %v", token.Claims)
}

// ForgetPassword godoc
// @Summary      Request password reset
// @Description  Generate a new password and send it via email to the param if the account exists
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.MailerRequest true "Email of the param"
// @Success      200 {object} nil "Success even if param does not exist (for security)"
// @Failure      400 {object} models.Error "Bad request (invalid JSON or email format)"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /forget-password [post]
func (s *Service) ForgetPassword(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	baseLogger := log.With().Logger()

	ctx := r.Context()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	var req models.MailerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		baseLogger.Error().Err(err).Msg("invalid JSON body")
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}

	logger := baseLogger.With().
		Str("email", req.Email).
		Logger()

	logger.Info().Msg("entering ForgetPassword")

	if _, err := mail.ParseAddress(req.Email); err != nil {
		logger.Warn().Err(err).Msg("invalid email address")
		return httpx.WriteError(w, http.StatusBadRequest, "Invalid email address")
	}

	user, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Error().Err(err).Msg("db get user by email failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	if user == nil {
		logger.Warn().Msg("user not found (masked success)")
		return httpx.Write(w, http.StatusOK, nil)
	}

	token, err := generateResetToken(req.Email)
	if err != nil {
		logger.Error().Err(err).Msg("reset token generation failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	logger.Debug().Int("token_len", len(token)).Msg("reset token generated")

	if err := s.mailer.SendLinkResetPassword(req.Email, "https://gfosd9euua.execute-api.eu-west-3.amazonaws.com/reset-password/"+token); err != nil {
		logger.Error().Err(err).Msg("sending reset link failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	logger.Info().Msg("reset link sent")
	return httpx.Write(w, http.StatusOK, nil)
}

// ResetPassword godoc
// @Summary      Réinitialise le mot de passe d’un utilisateur via un lien sécurisé
// @Description  Vérifie le token JWT fourni, génère un nouveau mot de passe, l'envoie par email à l'utilisateur et met à jour le mot de passe dans la base de données
// @Tags         auth
// @Accept       json
// @Produce      html
// @Param        token path string true "Token JWT de réinitialisation"
// @Success      200 {string} string "HTML indiquant que le mot de passe a été envoyé (même si l'utilisateur n'existe pas pour des raisons de sécurité)"
// @Failure      400 {object} models.Error "Requête invalide (ex: token manquant)"
// @Failure      500 {object} models.Error "Erreur interne du serveur (génération du mot de passe, envoi de mail, ou mise à jour en base)"
// @Router       /reset-password/{token} [get]
func (s *Service) ResetPassword(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	logger := log.With().Logger()

	token := chi.URLParam(r, "token")
	if token == "" {
		logger.Warn().Msg("missing reset token")
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}

	logger.Info().Msg("entering ResetPassword")

	email, err := parseResetToken(token)
	if err != nil {
		logger.Warn().Err(err).Msg("reset token invalid")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	newPassword, err := generatePassword(12)
	if err != nil {
		logger.Error().Err(err).Msg("password generation failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "Could not generate password")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("password hashing failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	if err := s.mailer.SendPasswordResetMail(email, newPassword); err != nil {
		logger.Error().Err(err).Str("email", email).Msg("sending new password failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	if err := s.db.ChangePassword(r.Context(), email, string(passwordHash)); err != nil {
		logger.Error().Err(err).Str("email", email).Msg("db change password failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	logger.Info().Str("email", email).Msg("password reset succeeded")
	return httpx.WriteHTMLResponse(w, http.StatusOK, "Mot de passe envoyé", "Un email contenant votre nouveau mot de passe vous a été envoyé à "+email+".")
}

// ChangePassword godoc
// @Summary      Change password for authenticated param
// @Description  Allows a connected param to change their password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.ChangePasswordRequest true "New password payload"
// @Success      200 {object} nil "Password changed successfully"
// @Failure      400 {object} models.Error "Bad request (invalid JSON)"
// @Failure      401 {object} models.Error "Unauthorized (not connected or param not found)"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /change-password [post]
func (s *Service) ChangePassword(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized: user not connected")
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}

	ctx := r.Context()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	user, err := s.db.GetUserById(ctx, ai.UserID)
	if err != nil {
		baseLogger.Error().Err(err).Str("user_id", ai.UserID).Msg("db get user by id failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}
	if user == nil {
		baseLogger.Warn().Str("user_id", ai.UserID).Msg("user not found")
		return httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		baseLogger.Error().Err(err).Str("user_id", ai.UserID).Msg("invalid JSON body")
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}

	logger := baseLogger.With().
		Str("user_id", ai.UserID).
		Bool("has_password", req.Password != "").
		Logger()

	logger.Info().Msg("entering ChangePassword")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("password hashing failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	if err := s.db.ChangePassword(ctx, user.Email, string(passwordHash)); err != nil {
		logger.Error().Err(err).Msg("db change password failed")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	logger.Info().Msg("password changed successfully")
	return httpx.Write(w, http.StatusOK, nil)
}
