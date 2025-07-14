package main

import (
	"PLIC/models"
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_GetMatchByID(t *testing.T) {
	court := models.NewDBCourtFixture()
	if court.Id == "" {
		court.Id = uuid.NewString()
	}

	match := models.NewDBMatchesFixture()
	match.CourtID = court.Id
	if match.Id == "" {
		match.Id = uuid.NewString()
	}

	user := models.NewDBUsersFixture()
	if user.Id == "" {
		user.Id = uuid.NewString()
	}

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
				Courts:  []models.DBCourt{court},
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
	court := models.NewDBCourtFixture()
	if court.Id == "" {
		court.Id = uuid.NewString()
	}
	match.CourtID = court.Id

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
				Courts:  []models.DBCourt{court},
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

func Test_GetMatchesByCourtId(t *testing.T) {
	court1 := models.NewDBCourtFixture()
	if court1.Id == "" {
		court1.Id = uuid.NewString()
	}
	court2 := models.NewDBCourtFixture()
	if court2.Id == "" {
		court2.Id = uuid.NewString()
	}

	match1 := models.DBMatches{
		Id:           uuid.NewString(),
		Sport:        models.Foot,
		Place:        "Paris",
		Date:         time.Now().Add(-time.Hour),
		CurrentState: models.Termine,
		Score1:       3,
		Score2:       2,
		CourtID:      court1.Id,
	}
	match2 := models.DBMatches{
		Id:           uuid.NewString(),
		Sport:        models.Basket,
		Place:        "Lyon",
		Date:         time.Now(),
		CurrentState: models.Termine,
		Score1:       1,
		Score2:       1,
		CourtID:      court2.Id,
	}

	testCases := []struct {
		name           string
		fixtures       DBFixtures
		courtID        string
		auth           models.AuthInfo
		expectedCode   int
		expectResponse bool
		expectErrorMsg string
	}{
		{
			name:           "Missing courtId in URL params",
			courtID:        "",
			auth:           models.AuthInfo{IsConnected: true},
			expectedCode:   http.StatusBadRequest,
			expectErrorMsg: "missing courtId in url params",
		},
		{
			name:           "User not connected",
			courtID:        court1.Id,
			auth:           models.AuthInfo{IsConnected: false},
			expectedCode:   http.StatusUnauthorized,
			expectErrorMsg: "not authorized",
		},
		{
			name: "No matches for given court",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match2},
			},
			courtID:        court1.Id,
			auth:           models.AuthInfo{IsConnected: true},
			expectedCode:   http.StatusNotFound,
			expectErrorMsg: "no matches found for this court",
		},
		{
			name: "Matches found for court1",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2},
			},
			courtID:        court1.Id,
			auth:           models.AuthInfo{IsConnected: true},
			expectedCode:   http.StatusOK,
			expectResponse: true,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			url := "/match/court/" + c.courtID
			r := httptest.NewRequest("GET", url, nil)

			routeCtx := chi.NewRouteContext()
			if c.courtID != "" {
				routeCtx.URLParams.Add("courtId", c.courtID)
			}
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()

			err := s.GetMatchesByCourtId(w, r, c.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, c.expectedCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if c.expectResponse {
				var matches []models.GetMatchByCourtIdResponses
				err = json.Unmarshal(body, &matches)
				require.NoError(t, err)
				require.NotEmpty(t, matches)
			} else {
				require.Contains(t, string(body), c.expectErrorMsg)
			}
		})
	}
}

func Test_GetAllMatches(t *testing.T) {
	match1 := models.NewDBMatchesFixture()
	match2 := models.NewDBMatchesFixture()

	court1 := models.NewDBCourtFixture()
	if court1.Id == "" {
		court1.Id = uuid.NewString()
	}
	court2 := models.NewDBCourtFixture()
	if court2.Id == "" {
		court2.Id = uuid.NewString()
	}

	match1.CourtID = court1.Id
	match2.CourtID = court2.Id

	user := models.NewDBUsersFixture()

	fixtures := DBFixtures{
		Courts:  []models.DBCourt{court1, court2},
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

	s := &Service{}
	s.InitServiceTest()
	s.loadFixtures(DBFixtures{
		Users: []models.DBUsers{user},
	})

	ctx := context.Background()

	court := models.NewDBCourtFixture().
		WithName("Test Court").
		WithLatitude(48.8566).
		WithLongitude(2.3522)

	err := s.db.InsertCourtForTest(ctx, court)
	require.NoError(t, err)

	matchReq := models.MatchRequest{
		Sport:           models.Foot,
		CourtID:         court.Id,
		Date:            time.Now().Add(24 * time.Hour),
		NbreParticipant: 6,
	}

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
	require.Equal(t, court.Address, res.Place)
	require.WithinDuration(t, matchReq.Date, res.Date, time.Second)
	require.Equal(t, matchReq.NbreParticipant, res.NbreParticipant)
	require.Equal(t, models.ManqueJoueur, res.CurrentState)
	require.Equal(t, 0, res.Score1)
	require.Equal(t, 0, res.Score2)
	require.Len(t, res.Users, 1)
	require.Equal(t, user.Username, res.Users[0].Username)
}

func Test_UpdateMatchScore(t *testing.T) {
	match := models.NewDBMatchesFixture()
	user := models.NewDBUsersFixture()

	court := models.NewDBCourtFixture()
	if court.Id == "" {
		court.Id = uuid.NewString()
	}
	match.CourtID = court.Id

	matchID := match.Id
	userID := user.Id

	updateReq := models.UpdateScoreRequest{
		Score1: 3,
		Score2: 2,
	}

	s := &Service{}
	s.InitServiceTest()
	s.loadFixtures(DBFixtures{
		Courts:  []models.DBCourt{court},
		Matches: []models.DBMatches{match},
		Users:   []models.DBUsers{user},
		UserMatches: []models.DBUserMatch{
			{
				UserID:    userID,
				MatchID:   matchID,
				CreatedAt: time.Now(),
			},
		},
	})

	bodyBytes, err := json.Marshal(updateReq)
	require.NoError(t, err)

	r := httptest.NewRequest("PATCH", "/score/match/"+matchID, bytes.NewReader(bodyBytes))
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", matchID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

	w := httptest.NewRecorder()
	err = s.UpdateMatchScore(w, r, models.AuthInfo{
		IsConnected: true,
		UserID:      userID,
	})
	require.NoError(t, err)

	resp := w.Result()
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var res models.MatchResponse
	err = json.Unmarshal(body, &res)
	require.NoError(t, err)

	require.Equal(t, updateReq.Score1, res.Score1)
	require.Equal(t, updateReq.Score2, res.Score2)
	require.Equal(t, matchID, res.Id)
}
