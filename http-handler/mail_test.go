package main

import (
	"PLIC/mailer"
	"PLIC/models"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_SendTestMail(t *testing.T) {
	type expected struct {
		statusCode int
		mailsSent  map[string]int
	}

	type testCase struct {
		name     string
		param    models.MailerRequest
		expected expected
	}

	testCases := []testCase{
		{
			name: "Basic test",
			param: models.MailerRequest{
				Email: "example@gmail.com",
			},
			expected: expected{
				statusCode: http.StatusOK,
				mailsSent: map[string]int{
					"test": 1,
				},
			},
		},
		{
			name: "Invalid email",
			param: models.MailerRequest{
				Email: "NotAnEmail",
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
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

			body, _ := json.Marshal(c.param)

			req := httptest.NewRequest("POST", "/email", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := s.SendMail(w, req, models.AuthInfo{})
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.statusCode, resp.StatusCode)
			if c.expected.statusCode != http.StatusOK {
				return
			}
			require.Equal(t, c.expected.mailsSent, mockMailer.SentCounts)
		})
	}
}
