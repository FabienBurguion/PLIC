package main

import (
	"PLIC/models"
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_GetMatchByID(t *testing.T) {
	match := models.NewDBMatchesFixture()
	user := models.NewDBUsersFixture()
	matchID := match.Id

	type testCase struct {
		name         string
		fixtures     DBFixtures
		paramID      string
		expectedCode int
		expectFound  bool
	}

	testCases := []testCase{
		{
			name: "Match found",
			fixtures: DBFixtures{
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					{
						UserID:    user.Id,
						MatchID:   matchID,
						CreatedAt: time.Now(),
					},
				},
			},
			paramID:      matchID,
			expectedCode: 200,
			expectFound:  true,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			r := httptest.NewRequest("GET", "/match/"+c.paramID, nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", c.paramID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			err := s.GetMatchByID(w, r, models.AuthInfo{
				IsConnected: true,
				UserID:      user.Id,
			})
			require.NoError(t, err)

			resp := w.Result()
			defer resp.Body.Close()
			require.Equal(t, c.expectedCode, resp.StatusCode)

			if c.expectFound {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				var res models.MatchResponse
				err = json.Unmarshal(body, &res)
				require.NoError(t, err)
				require.Equal(t, match.Id, res.Id)
			}
		})
	}
}

func Test_GetMatchesByUserID(t *testing.T) {
	match := models.NewDBMatchesFixture()
	user := models.NewDBUsersFixture()
	matchID := match.Id
	userID := user.Id

	type testCase struct {
		name         string
		fixtures     DBFixtures
		paramID      string
		expectedCode int
		expectFound  bool
	}

	testCases := []testCase{
		{
			name: "Matches found for user",
			fixtures: DBFixtures{
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					{
						UserID:    userID,
						MatchID:   matchID,
						CreatedAt: time.Now(),
					},
				},
			},
			paramID:      userID,
			expectedCode: 200,
			expectFound:  true,
		},
		{
			name: "No matches for user",
			fixtures: DBFixtures{
				Users: []models.DBUsers{user},
			},
			paramID:      userID,
			expectedCode: 404,
			expectFound:  false,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			r := httptest.NewRequest("GET", "/user/matches/"+c.paramID, nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("userId", c.paramID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			err := s.GetMatchesByUserID(w, r, models.AuthInfo{
				IsConnected: true,
				UserID:      userID,
			})
			require.NoError(t, err)

			resp := w.Result()
			defer resp.Body.Close()
			require.Equal(t, c.expectedCode, resp.StatusCode)

			if c.expectFound {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				var res []models.GetMatchByUserIdResponses
				err = json.Unmarshal(body, &res)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				require.Equal(t, matchID, res[0].Id)
			}
		})
	}
}

func Test_GetAllMatches(t *testing.T) {
	match1 := models.NewDBMatchesFixture()
	match2 := models.NewDBMatchesFixture()
	user := models.NewDBUsersFixture()

	fixtures := DBFixtures{
		Matches: []models.DBMatches{match1, match2},
		Users:   []models.DBUsers{user},
		UserMatches: []models.DBUserMatch{
			{UserID: user.Id, MatchID: match1.Id, CreatedAt: time.Now()},
			{UserID: user.Id, MatchID: match2.Id, CreatedAt: time.Now()},
		},
	}

	s := &Service{}
	s.InitServiceTest()
	s.loadFixtures(fixtures)

	r := httptest.NewRequest("GET", "/match/all", nil)
	w := httptest.NewRecorder()

	err := s.GetAllMatches(w, r, models.AuthInfo{
		IsConnected: true,
		UserID:      user.Id,
	})
	require.NoError(t, err)

	resp := w.Result()
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var res []models.MatchResponse
	err = json.Unmarshal(body, &res)
	require.NoError(t, err)

	require.Len(t, res, 2)
}

func Test_CreateMatch(t *testing.T) {
	user := models.NewDBUsersFixture()
	matchReq := models.MatchRequest{
		Sport: models.Foot,
		Place: "Lyon",
		Date:  time.Now().Add(24 * time.Hour),
	}

	s := &Service{}
	s.InitServiceTest()
	s.loadFixtures(DBFixtures{
		Users: []models.DBUsers{user},
	})

	bodyBytes, err := json.Marshal(matchReq)
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/match", io.NopCloser(bytes.NewReader(bodyBytes)))
	w := httptest.NewRecorder()

	err = s.CreateMatch(w, r, models.AuthInfo{
		IsConnected: true,
		UserID:      user.Id,
	})
	require.NoError(t, err)

	resp := w.Result()
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var res models.MatchResponse
	err = json.Unmarshal(body, &res)
	require.NoError(t, err)

	require.Equal(t, matchReq.Sport, res.Sport)
	require.Equal(t, matchReq.Place, res.Place)
	require.WithinDuration(t, matchReq.Date, res.Date, time.Second)
	require.Equal(t, 1, res.ParticipantNber)
	require.Equal(t, models.ManqueJoueur, res.CurrentState)
	require.Equal(t, 0, res.Score1)
	require.Equal(t, 0, res.Score2)
	require.Len(t, res.Users, 1)
	require.Equal(t, user.Username, res.Users[0].Username)
}
