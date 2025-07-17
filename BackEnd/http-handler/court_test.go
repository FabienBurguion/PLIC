package main

import (
	"PLIC/models"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_GetCourtByID(t *testing.T) {
	type testCase struct {
		name          string
		urlID         string
		fixtures      DBFixtures
		expectedCode  int
		expectedCheck func(t *testing.T, res models.DBCourt)
	}

	court := models.NewDBCourtFixture()

	testCases := []testCase{
		{
			name:  "Court found",
			urlID: court.Id,
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
			},
			expectedCode: http.StatusOK,
			expectedCheck: func(t *testing.T, res models.DBCourt) {
				require.Equal(t, court.Id, res.Id)
				require.Equal(t, court.Name, res.Name)
				require.Equal(t, court.Address, res.Address)
				require.Equal(t, court.Latitude, res.Latitude)
				require.Equal(t, court.Longitude, res.Longitude)
			},
		},
		{
			name:          "Missing ID",
			urlID:         "",
			fixtures:      DBFixtures{},
			expectedCode:  http.StatusBadRequest,
			expectedCheck: nil,
		},
		{
			name:          "Court not found",
			urlID:         uuid.NewString(),
			fixtures:      DBFixtures{},
			expectedCode:  http.StatusNotFound,
			expectedCheck: nil,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			url := "/court"
			if c.urlID != "" {
				url += "/" + c.urlID
			}

			r := httptest.NewRequest("GET", url, nil)
			if c.urlID != "" {
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("id", c.urlID)
				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
			}

			w := httptest.NewRecorder()
			err := s.GetCourtByID(w, r, models.AuthInfo{})
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			require.Equal(t, c.expectedCode, resp.StatusCode)

			if c.expectedCode == http.StatusOK && c.expectedCheck != nil {
				var res models.DBCourt
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				err = json.Unmarshal(body, &res)
				require.NoError(t, err)

				c.expectedCheck(t, res)
			}
		})
	}
}
