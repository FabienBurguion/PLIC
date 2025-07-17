package main

import (
	"PLIC/mailer"
	"PLIC/models"
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestService_Login(t *testing.T) {
	type expected struct {
		code     int
		response models.LoginResponse
	}
	type testCase struct {
		name     string
		fixtures DBFixtures
		param    models.LoginRequest
		expected expected
	}

	userId := uuid.NewString()
	username := "Username"
	password := "Password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     tokenDuration,
		"iat":     time.Now().Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tok.SignedString([]byte(jwtSecret))
	require.NoError(t, err)

	testCases := []testCase{
		{
			name: "User exists -> can login and receives token",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithUsername(username).
						WithPassword(string(hashedPassword)),
				},
			},
			param: models.LoginRequest{
				Username: username,
				Password: password,
			},
			expected: expected{
				code:     http.StatusOK,
				response: models.LoginResponse{Token: token},
			},
		},
		{
			name: "Wrong username => 401",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithUsername(username).
						WithPassword(string(hashedPassword)),
				},
			},
			param: models.LoginRequest{
				Username: "Another username",
				Password: password,
			},
			expected: expected{
				code: http.StatusUnauthorized,
			},
		},
		{
			name: "Wrong password => 401",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithUsername(username).
						WithPassword(string(hashedPassword)),
				},
			},
			param: models.LoginRequest{
				Username: username,
				Password: "Another password",
			},
			expected: expected{
				code: http.StatusUnauthorized,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			body, _ := json.Marshal(c.param)

			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			err := s.Login(w, req, models.AuthInfo{})
			require.NoError(t, err)
			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.code, resp.StatusCode)
			if c.expected.code != http.StatusOK {
				return
			}
			var actual models.LoginResponse
			err = json.NewDecoder(resp.Body).Decode(&actual)
			require.NoError(t, err)
			parsedToken, err := jwt.Parse(actual.Token, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			require.NoError(t, err)
			require.True(t, parsedToken.Valid)

			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			require.True(t, ok)
			require.Equal(t, userId, claims["user_id"])
		})
	}
}

func TestService_Register(t *testing.T) {
	type expected struct {
		code int
	}
	type testCase struct {
		name     string
		fixtures DBFixtures
		param    models.RegisterRequest
		expected expected
	}

	userId := uuid.NewString()
	email := "NewEmail"
	password := "NewPassword"

	testCases := []testCase{
		{
			name: "User does not exist -> can register and receives token",
			fixtures: DBFixtures{
				Users: []models.DBUsers{},
			},
			param: models.RegisterRequest{
				Email:    email,
				Password: password,
			},
			expected: expected{
				code: http.StatusCreated,
			},
		},
		{
			name: "Empty email => bad request",
			fixtures: DBFixtures{
				Users: []models.DBUsers{},
			},
			param: models.RegisterRequest{
				Email:    "",
				Password: password,
			},
			expected: expected{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Empty password => bad request",
			fixtures: DBFixtures{
				Users: []models.DBUsers{},
			},
			param: models.RegisterRequest{
				Email:    email,
				Password: "",
			},
			expected: expected{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "User already exists -> 401",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithEmail(email),
				},
			},
			param: models.RegisterRequest{
				Email:    email,
				Password: password,
			},
			expected: expected{
				code: http.StatusUnauthorized,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			body, _ := json.Marshal(c.param)

			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := s.Register(w, req, models.AuthInfo{})
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.code, resp.StatusCode)

			if c.expected.code != http.StatusCreated {
				return
			}

			var actual models.LoginResponse
			err = json.NewDecoder(resp.Body).Decode(&actual)
			require.NoError(t, err)

			parsedToken, err := jwt.Parse(actual.Token, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			require.NoError(t, err)
			require.True(t, parsedToken.Valid)

			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			require.True(t, ok)
			require.NotEmpty(t, claims["user_id"])
		})
	}
}

func TestService_ForgetPassword(t *testing.T) {
	type expected struct {
		statusCode int
		mailSent   map[string]int
	}

	type testCase struct {
		name     string
		param    models.MailerRequest
		fixtures DBFixtures
		expected expected
	}

	userId := uuid.NewString()

	testCases := []testCase{
		{
			name: "Password changes + mail sent",
			param: models.MailerRequest{
				Email: "example@gmail.com",
			},
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithEmail("example@gmail.com"),
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				mailSent: map[string]int{
					"link_reset": 1,
				},
			},
		},
		{
			name: "Not an email -> 400",
			param: models.MailerRequest{
				Email: "notAnEmail",
			},
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithEmail("example@gmail.com"),
				},
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "param does not exist",
			param: models.MailerRequest{
				Email: "example@gmail.com",
			},
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture(),
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				mailSent:   map[string]int{},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			mockMailer := mailer.NewMockMailer()
			s.mailer = mockMailer
			s.loadFixtures(c.fixtures)

			body, _ := json.Marshal(c.param)

			req := httptest.NewRequest("POST", "/forget-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := s.ForgetPassword(w, req, models.AuthInfo{})
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.statusCode, resp.StatusCode)
			if c.expected.statusCode != http.StatusOK {
				return
			}
			require.Equal(t, c.expected.mailSent, mockMailer.SentCounts)
		})
	}
}

func TestService_ResetPassword(t *testing.T) {
	type expected struct {
		statusCode int
		mailSent   map[string]int
	}

	type testCase struct {
		name     string
		param    models.DBUsers
		expected expected
	}

	userEmail := "reset@example.com"
	userId := uuid.NewString()

	testCases := []testCase{
		{
			name: "Valid token, password updated, mail sent",
			param: models.NewDBUsersFixture().
				WithId(userId).
				WithEmail(userEmail).
				WithPassword("old-password-hash"),
			expected: expected{
				statusCode: http.StatusOK,
				mailSent: map[string]int{
					"password_forgot": 1,
				},
			},
		},
		{
			name: "Missing token → 400",
			param: models.NewDBUsersFixture().
				WithId(userId).
				WithEmail(userEmail),
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Invalid token → 500",
			param: models.NewDBUsersFixture().
				WithId(userId).
				WithEmail(userEmail),
			expected: expected{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			mockMailer := mailer.NewMockMailer()
			s.mailer = mockMailer
			s.loadFixtures(DBFixtures{
				Users: []models.DBUsers{c.param},
			})

			var token string
			var err error
			switch c.name {
			case "Valid token, password updated, mail sent":
				token, err = GenerateResetToken(userEmail)
				require.NoError(t, err)
			case "Invalid token → 500":
				token = "invalid.token.value"
			case "Missing token → 400":
				token = ""
			}

			req := httptest.NewRequest("GET", "/reset-password/"+token, nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("token", token)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			w := httptest.NewRecorder()

			err = s.ResetPassword(w, req, models.AuthInfo{})
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.statusCode, resp.StatusCode)

			if c.expected.statusCode != http.StatusOK {
				return
			}

			require.Equal(t, c.expected.mailSent, mockMailer.SentCounts)

			updatedUser, err := s.db.GetUserByEmail(req.Context(), userEmail)
			require.NoError(t, err)
			require.NotEqual(t, c.param.Password, updatedUser.Password)
		})
	}
}

func TestService_ChangePassword(t *testing.T) {
	type expected struct {
		statusCode      int
		currentPassword string
	}

	type testCase struct {
		name     string
		param    models.AuthInfo
		fixtures DBFixtures
		expected expected
	}

	userId := uuid.NewString()
	oldPassword := "oldSecurePassword"
	newPassword := "newSecurePassword"

	param := models.ChangePasswordRequest{
		Password: newPassword,
	}

	testCases := []testCase{
		{
			name: "User does not exist -> 401",
			param: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			expected: expected{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "User exists -> 200",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithPassword(oldPassword),
				},
			},
			param: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			expected: expected{
				statusCode:      http.StatusOK,
				currentPassword: newPassword,
			},
		},
		{
			name: "Not authentified -> 401",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithPassword(oldPassword),
				},
			},
			param: models.AuthInfo{
				IsConnected: false,
			},
			expected: expected{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Incorrect user Id -> 401",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithPassword(oldPassword),
				},
			},
			param: models.AuthInfo{
				IsConnected: true,
				UserID:      uuid.NewString(),
			},
			expected: expected{
				statusCode: http.StatusUnauthorized,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			body, _ := json.Marshal(param)

			req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := s.ChangePassword(w, req, c.param)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.statusCode, resp.StatusCode)
			if c.expected.statusCode != http.StatusOK {
				return
			}
			user, err := s.db.GetUserById(ctx, userId)
			require.NoError(t, err)
			require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(c.expected.currentPassword)))
		})
	}
}
