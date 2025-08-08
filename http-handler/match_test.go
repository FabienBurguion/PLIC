package main

import (
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
	"github.com/stretchr/testify/require"
)

func Test_GetMatchByID(t *testing.T) {
	type expected struct {
		code  int
		found bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected expected
	}
	court := models.NewDBCourtFixture()

	match := models.NewDBMatchesFixture().
		WithCourtId(court.Id)

	user := models.NewDBUsersFixture()

	testCases := []testCase{
		{
			name: "Match found",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user.Id).
						WithMatchId(match.Id),
				},
			},
			param: match.Id,
			expected: expected{
				code:  200,
				found: true,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(c.fixtures)

			r := httptest.NewRequest("GET", "/match/"+c.param, nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", c.param)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			err := s.GetMatchByID(w, r, models.AuthInfo{
				IsConnected: true,
				UserID:      user.Id,
			})
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.code, resp.StatusCode)

			if c.expected.found {
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
	type expected struct {
		code  int
		found bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected expected
	}

	court := models.NewDBCourtFixture()
	match := models.NewDBMatchesFixture().
		WithCourtId(court.Id)

	user := models.NewDBUsersFixture()

	testCases := []testCase{
		{
			name: "Matches found for user",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user.Id).
						WithMatchId(match.Id),
				},
			},
			param: user.Id,
			expected: expected{
				code:  200,
				found: true,
			},
		},
		{
			name: "No matches for user",
			fixtures: DBFixtures{
				Users: []models.DBUsers{user},
			},
			param: user.Id,
			expected: expected{
				code:  200,
				found: false,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(c.fixtures)

			r := httptest.NewRequest("GET", "/user/matches/"+c.param, nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("userId", c.param)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			err := s.GetMatchesByUserID(w, r, models.AuthInfo{
				IsConnected: true,
				UserID:      c.param,
			})
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.code, resp.StatusCode)

			if c.expected.found {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				var res []models.GetMatchByUserIdResponses
				err = json.Unmarshal(body, &res)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				require.Equal(t, match.Id, res[0].Id)
			}
		})
	}
}

func Test_GetMatchesByCourtId(t *testing.T) {
	type expected struct {
		code     int
		response bool
		errorMsg string
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		auth     models.AuthInfo
		expected expected
	}

	court1 := models.NewDBCourtFixture()
	court2 := models.NewDBCourtFixture()

	match1 := models.NewDBMatchesFixture().
		WithCourtId(court1.Id).
		WithSport(models.Foot).
		WithCurrentState(models.Termine)
	match2 := models.NewDBMatchesFixture().
		WithCourtId(court2.Id).
		WithSport(models.Basket).
		WithCurrentState(models.Termine)

	testCases := []testCase{
		{
			name:  "Missing courtId in URL params",
			param: "",
			auth:  models.AuthInfo{IsConnected: true},
			expected: expected{
				code:     http.StatusBadRequest,
				errorMsg: "missing courtId in url params",
			},
		},
		{
			name:  "User not connected",
			param: court1.Id,
			auth:  models.AuthInfo{IsConnected: false},
			expected: expected{
				code:     http.StatusUnauthorized,
				errorMsg: "not authorized",
			},
		},
		{
			name: "No matches for given court",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match2},
			},
			param: court1.Id,
			auth:  models.AuthInfo{IsConnected: true},
			expected: expected{
				code:     http.StatusOK,
				response: true,
			},
		},
		{
			name: "Matches found for court1",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2},
			},
			param: court1.Id,
			auth:  models.AuthInfo{IsConnected: true},
			expected: expected{
				code:     http.StatusOK,
				response: true,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(c.fixtures)

			url := "/match/court/" + c.param
			r := httptest.NewRequest("GET", url, nil)

			routeCtx := chi.NewRouteContext()
			if c.param != "" {
				routeCtx.URLParams.Add("courtId", c.param)
			}
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()

			err := s.GetMatchesByCourtId(w, r, c.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			require.Equal(t, c.expected.code, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if c.expected.response {
				var matches []models.MatchResponse
				err = json.Unmarshal(body, &matches)
				require.NoError(t, err)
			} else {
				require.Contains(t, string(body), c.expected.errorMsg)
			}
		})
	}
}

func Test_GetAllMatches(t *testing.T) {
	type expected struct {
		code   int
		length int
	}

	type testCase struct {
		name     string
		auth     models.AuthInfo
		fixtures DBFixtures
		expected expected
	}

	court1 := models.NewDBCourtFixture()
	court2 := models.NewDBCourtFixture()

	match1 := models.NewDBMatchesFixture().
		WithCourtId(court1.Id)
	match2 := models.NewDBMatchesFixture().
		WithCourtId(court2.Id)

	user := models.NewDBUsersFixture()

	testCases := []testCase{
		{
			name: "User connected with 2 matches",
			auth: models.AuthInfo{IsConnected: true, UserID: user.Id},
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user.Id).
						WithMatchId(match1.Id),
					models.NewDBUserMatchFixture().
						WithUserId(user.Id).
						WithMatchId(match2.Id),
				},
			},
			expected: expected{
				code:   http.StatusOK,
				length: 2,
			},
		},
		{
			name: "No matches",
			auth: models.AuthInfo{IsConnected: true, UserID: user.Id},
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court1, court2},
				Users:  []models.DBUsers{user},
			},
			expected: expected{
				code:   http.StatusOK,
				length: 0,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(c.fixtures)

			r := httptest.NewRequest("GET", "/match/all", nil)
			w := httptest.NewRecorder()

			err := s.GetAllMatches(w, r, c.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.code, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var res []models.MatchResponse
			err = json.Unmarshal(body, &res)
			require.NoError(t, err)
			require.Len(t, res, c.expected.length)
		})
	}
}

func Test_CreateMatch(t *testing.T) {
	type expected struct {
		statusCode int
	}

	type testCase struct {
		name        string
		fixtures    DBFixtures
		auth        models.AuthInfo
		insertCourt bool
		param       models.MatchRequest
		expected    expected
	}
	user := models.NewDBUsersFixture()
	court := models.NewDBCourtFixture().
		WithName("Test Court").
		WithLatitude(48.8566).
		WithLongitude(2.3522)

	testCases := []testCase{
		{
			name: "Successful match creation",
			auth: models.AuthInfo{IsConnected: true, UserID: user.Id},
			fixtures: DBFixtures{
				Users: []models.DBUsers{user},
			},
			insertCourt: true,
			param: models.NewMatchRequestFixture().
				WithCourtId(court.Id),
			expected: expected{
				statusCode: http.StatusCreated,
			},
		},
		{
			name: "Unauthorized user",
			auth: models.AuthInfo{IsConnected: false, UserID: user.Id},
			fixtures: DBFixtures{
				Users: []models.DBUsers{user},
			},
			insertCourt: true,
			param: models.NewMatchRequestFixture().
				WithCourtId(court.Id),
			expected: expected{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Court does not exist",
			auth: models.AuthInfo{IsConnected: true, UserID: user.Id},
			fixtures: DBFixtures{
				Users: []models.DBUsers{user},
			},
			insertCourt: false,
			param: models.NewMatchRequestFixture().
				WithCourtId(court.Id),
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(c.fixtures)

			if c.insertCourt {
				err := s.db.InsertCourtForTest(context.Background(), court)
				require.NoError(t, err)
			}

			bodyBytes, err := json.Marshal(c.param)
			require.NoError(t, err)

			r := httptest.NewRequest("POST", "/match", io.NopCloser(bytes.NewReader(bodyBytes)))
			w := httptest.NewRecorder()

			err = s.CreateMatch(w, r, c.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.statusCode, resp.StatusCode)
		})
	}
}

func Test_UpdateMatchScore(t *testing.T) {
	type testCase struct {
		name      string
		fixtures  DBFixtures
		auth      models.AuthInfo
		param     string
		updateReq models.UpdateScoreRequest
		expected  int
	}

	user := models.NewDBUsersFixture()
	court := models.NewDBCourtFixture()
	match := models.NewDBMatchesFixture().
		WithCourtId(court.Id)

	testCases := []testCase{
		{
			name:  "Successful score update",
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			param: match.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user.Id).
						WithMatchId(match.Id),
				},
			},
			updateReq: models.NewUpdateScoreRequestFixture().
				WithScore1(3).
				WithScore2(2),
			expected: http.StatusOK,
		},
		{
			name:  "Missing match ID param",
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			param: "",
			updateReq: models.NewUpdateScoreRequestFixture().
				WithScore1(3).
				WithScore2(2),
			expected: http.StatusBadRequest,
		},
		{
			name:     "Unauthorized user",
			auth:     models.AuthInfo{IsConnected: false, UserID: user.Id},
			param:    match.Id,
			expected: http.StatusUnauthorized,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(c.fixtures)

			bodyBytes, err := json.Marshal(c.updateReq)
			require.NoError(t, err)

			url := "/score/match/" + c.param
			r := httptest.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))

			routeCtx := chi.NewRouteContext()
			if c.param != "" {
				routeCtx.URLParams.Add("id", c.param)
			}
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()

			err = s.UpdateMatchScore(w, r, c.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected, resp.StatusCode)
		})
	}
}

func Test_JoinMatch(t *testing.T) {
	type expected struct {
		bodyJSON    string
		code        int
		errorMsg    string
		checkJoined bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		auth     models.AuthInfo
		expected expected
	}

	court := models.NewDBCourtFixture().
		WithAddress("123 Rue Sport")

	match := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithParticipantNber(2).
		WithCurrentState(models.ManqueJoueur)

	user := models.NewDBUsersFixture()
	teammate := models.NewDBUsersFixture().
		WithUsername("teammate").
		WithEmail("other_email@gmail.com")

	testCases := []testCase{
		{
			name: "User successfully joins match",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user},
			},
			param: match.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			expected: expected{
				bodyJSON:    `{"team": 1}`,
				code:        http.StatusOK,
				checkJoined: true,
			},
		},
		{
			name:  "Missing match ID",
			param: "",
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			expected: expected{
				bodyJSON: `{"team": 1}`,
				code:     http.StatusBadRequest,
				errorMsg: "missing match ID",
			},
		},
		{
			name:  "Unauthorized user",
			param: match.Id,
			auth:  models.AuthInfo{IsConnected: false, UserID: user.Id},
			expected: expected{
				bodyJSON: `{"team": 1}`,
				code:     http.StatusUnauthorized,
				errorMsg: "not authorized",
			},
		},
		{
			name: "User already joined",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user.Id).
						WithMatchId(match.Id),
				},
			},
			param: match.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			expected: expected{
				bodyJSON: `{"team": 1}`,
				code:     http.StatusConflict,
				errorMsg: "user already joined",
			},
		},
		{
			name: "Team is full",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user, teammate},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(teammate.Id).
						WithMatchId(match.Id),
				},
			},
			param: match.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			expected: expected{
				bodyJSON: `{"team": 1}`,
				code:     http.StatusBadRequest,
				errorMsg: "this team is full",
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(c.fixtures)

			url := "/match/join/" + c.param
			r := httptest.NewRequest("POST", url, strings.NewReader(c.expected.bodyJSON))
			r.Header.Set("Content-Type", "application/json")

			routeCtx := chi.NewRouteContext()
			if c.param != "" {
				routeCtx.URLParams.Add("id", c.param)
			}
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()

			err := s.JoinMatch(w, r, c.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			require.Equal(t, c.expected.code, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if c.expected.errorMsg != "" {
				require.Contains(t, string(body), c.expected.errorMsg)
			}

			if c.expected.checkJoined {
				ctx := context.Background()

				joined, err := s.db.IsUserInMatch(ctx, user.Id, match.Id)
				require.NoError(t, err)
				require.True(t, joined)

				updatedMatch, err := s.db.GetMatchById(ctx, match.Id)
				require.NoError(t, err)
				require.Equal(t, models.Valide, updatedMatch.CurrentState)
			}
		})
	}
}
