package main

import (
	"PLIC/models"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloWorldHandler(t *testing.T) {
	type expected struct {
		statusCode int
		response   models.HelloWorldResponse
	}

	type testCase struct {
		name     string
		param    string
		expected expected
	}

	testCases := []testCase{
		{
			name:  "Basic test",
			param: "Emma",
			expected: expected{
				statusCode: http.StatusOK,
				response:   models.HelloWorldResponse{Response: "Hello Emma"},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/hello_world", nil)
			q := r.URL.Query()
			q.Set("name", c.param)
			r.URL.RawQuery = q.Encode()

			err := s.GetHelloWorld(w, r, models.AuthInfo{})
			require.NoError(t, err)
			require.Equal(t, c.expected.statusCode, w.Code)
			var actualBody models.HelloWorldResponse
			err = json.Unmarshal(w.Body.Bytes(), &actualBody)
			require.NoError(t, err)
			require.Equal(t, c.expected.response, actualBody)
		})
	}
}
