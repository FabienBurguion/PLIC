package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_POCDB(t *testing.T) {
	type testCase struct {
		name  string
		param string
	}

	id := uuid.NewString()

	testCases := []testCase{
		{
			name:  "Basic test",
			param: id,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()

			ctx := context.Background()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "", nil)
			q := r.URL.Query()
			q.Set("id", c.param)
			r.URL.RawQuery = q.Encode()

			err := s.CreateUser(w, r)
			require.NoError(t, err)
			user, err := s.db.GetUser(ctx, id)
			require.NoError(t, err)
			require.Equal(t, user.Id, c.param)
		})
	}
}
