package main

import (
	"PLIC/mailer"
	"PLIC/models"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
		statusCode int
		persist    bool
		body       *string
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    models.RegisterRequest
		expected expected
	}

	email := "new@gmail.com"
	password := "NewPassword"

	tcs := []testCase{
		{
			name: "User does not exist -> can register and receives token (no username/bio provided)",
			fixtures: DBFixtures{
				Users: []models.DBUsers{},
			},
			param: models.RegisterRequest{
				Email:    email,
				Password: password,
				Username: "user1",
			},
			expected: expected{
				statusCode: http.StatusCreated,
				persist:    true,
			},
		},
		{
			name: "User does not exist -> can register and receives token (with username & bio)",
			fixtures: DBFixtures{
				Users: []models.DBUsers{},
			},
			param: models.RegisterRequest{
				Email:    "with-username@example.com",
				Password: "pwd",
				Username: "Neo",
				Bio:      ptr("The One"),
			},
			expected: expected{
				statusCode: http.StatusCreated,
				persist:    true,
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
				statusCode: http.StatusBadRequest,
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
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Email already exists => 409 email_taken",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(uuid.NewString()).
						WithEmail("dup@example.com").
						WithUsername("someone"),
				},
			},
			param: models.RegisterRequest{
				Email:    "dup@example.com",
				Password: "pwd",
				Username: "anyone",
			},
			expected: expected{
				statusCode: http.StatusConflict,
				body:       ptr(`{"message":"email_taken"}`),
			},
		},
		{
			name: "Username already exists => 409 username_taken",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(uuid.NewString()).
						WithEmail("existing@example.com").
						WithUsername("takenName"),
				},
			},
			param: models.RegisterRequest{
				Email:    "new-email-for-username-conflict@example.com",
				Password: "pwd",
				Username: "takenName",
			},
			expected: expected{
				statusCode: http.StatusConflict,
				body:       ptr(`{"message":"username_taken"}`),
			},
		},
	}

	for _, c := range tcs {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
			require.Equal(t, c.expected.statusCode, resp.StatusCode)

			if c.expected.body != nil {
				respBytes, _ := io.ReadAll(resp.Body)
				require.Equal(t, *c.expected.body, strings.TrimSpace(string(respBytes)))
			}

			if c.expected.statusCode != http.StatusCreated {
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

			if c.expected.persist {
				ctx := context.Background()
				u, err := s.db.GetUserByEmail(ctx, c.param.Email)
				require.NoError(t, err)
				require.NotNil(t, u)

				require.Equal(t, c.param.Username, u.Username)

				require.Equal(t, c.param.Bio, u.Bio)

				require.NotEmpty(t, u.Password)
				require.NotEqual(t, c.param.Password, u.Password)

				require.False(t, u.CreatedAt.IsZero())
				require.False(t, u.UpdatedAt.IsZero())
				require.NotEmpty(t, u.Id)
			}
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			mockMailer := mailer.NewMockMailer()
			s.mailer = mockMailer
			s.loadFixtures(DBFixtures{
				Users: []models.DBUsers{c.param},
			})

			var token string
			var err error
			switch c.name {
			case "Valid token, password updated, mail sent":
				token, err = generateResetToken(userEmail)
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
		statusCode  int
		newPassword string
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
				statusCode:  http.StatusOK,
				newPassword: newPassword,
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
			require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(c.expected.newPassword)))
		})
	}
}
