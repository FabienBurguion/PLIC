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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
		statusCode   int
		checkRanking bool
		wantElo      *int
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

	defaultElo := 1000
	existingElo := 1350

	testCases := []testCase{
		{
			name: "Successful match creation (creates default ranking if missing)",
			auth: models.AuthInfo{IsConnected: true, UserID: user.Id},
			fixtures: DBFixtures{
				Users: []models.DBUsers{user},
			},
			insertCourt: true,
			param: models.NewMatchRequestFixture().
				WithCourtId(court.Id),
			expected: expected{
				statusCode:   http.StatusCreated,
				checkRanking: true,
				wantElo:      &defaultElo,
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
		{
			name: "Successful match creation (keeps existing ranking as-is)",
			auth: models.AuthInfo{IsConnected: true, UserID: user.Id},
			fixtures: DBFixtures{
				Users:  []models.DBUsers{user},
				Courts: []models.DBCourt{court},
				Rankings: []models.DBRanking{
					{
						UserID:    user.Id,
						CourtID:   court.Id,
						Elo:       existingElo,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
			},
			insertCourt: false,
			param: models.NewMatchRequestFixture().
				WithCourtId(court.Id),
			expected: expected{
				statusCode:   http.StatusCreated,
				checkRanking: true,
				wantElo:      &existingElo,
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

			if c.expected.checkRanking && c.expected.statusCode == http.StatusCreated {
				ctx := context.Background()
				rk, err := s.db.GetRankingByUserAndCourt(ctx, user.Id, c.param.CourtID)
				require.NoError(t, err)
				require.NotNil(t, rk, "ranking should exist after CreateMatch")
				if c.expected.wantElo != nil {
					require.Equal(t, *c.expected.wantElo, rk.Elo)
				}
			}
		})
	}
}

func Test_UpdateMatchScore_ConsensusAndTeamRules(t *testing.T) {
	type expected struct {
		code          int
		state         models.MatchState
		score1        int
		score2        int
		errorContains string
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		steps    []struct {
			auth  models.AuthInfo
			param string
			req   models.UpdateScoreRequest
			exp   expected
		}
	}

	court := models.NewDBCourtFixture()
	matchConsensus := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCurrentState(models.ManqueScore)
	matchNoConsensus := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCurrentState(models.ManqueScore)
	matchTeamTwice := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCurrentState(models.ManqueScore)

	userA := models.NewDBUsersFixture()
	userB := models.NewDBUsersFixture().WithUsername("userB").WithEmail("emailB@example.com")
	userC := models.NewDBUsersFixture().WithUsername("userC").WithEmail("emailC@example.com")

	fixtures := DBFixtures{
		Courts:  []models.DBCourt{court},
		Matches: []models.DBMatches{matchConsensus, matchNoConsensus, matchTeamTwice},
		Users:   []models.DBUsers{userA, userB, userC},
		UserMatches: []models.DBUserMatch{
			models.NewDBUserMatchFixture().WithUserId(userA.Id).WithMatchId(matchConsensus.Id).WithTeam(1),
			models.NewDBUserMatchFixture().WithUserId(userB.Id).WithMatchId(matchConsensus.Id).WithTeam(2),

			models.NewDBUserMatchFixture().WithUserId(userA.Id).WithMatchId(matchNoConsensus.Id).WithTeam(1),
			models.NewDBUserMatchFixture().WithUserId(userB.Id).WithMatchId(matchNoConsensus.Id).WithTeam(2),

			models.NewDBUserMatchFixture().WithUserId(userA.Id).WithMatchId(matchTeamTwice.Id).WithTeam(1),
			models.NewDBUserMatchFixture().WithUserId(userC.Id).WithMatchId(matchTeamTwice.Id).WithTeam(1),
		},
	}

	testCases := []testCase{
		{
			name:     "Consensus: two teams same score -> match becomes Termine and ELO updates",
			fixtures: fixtures,
			steps: []struct {
				auth  models.AuthInfo
				param string
				req   models.UpdateScoreRequest
				exp   expected
			}{
				{
					auth:  models.AuthInfo{IsConnected: true, UserID: userA.Id},
					param: matchConsensus.Id,
					req:   models.NewUpdateScoreRequestFixture().WithScore1(3).WithScore2(2),
					exp:   expected{code: http.StatusOK, state: models.ManqueScore, score1: 3, score2: 2},
				},
				{
					auth:  models.AuthInfo{IsConnected: true, UserID: userB.Id},
					param: matchConsensus.Id,
					req:   models.NewUpdateScoreRequestFixture().WithScore1(3).WithScore2(2),
					exp:   expected{code: http.StatusOK, state: models.Termine, score1: 3, score2: 2},
				},
			},
		},
		{
			name:     "No consensus: second team different score -> state unchanged, score is last proposed, no ELO update",
			fixtures: fixtures,
			steps: []struct {
				auth  models.AuthInfo
				param string
				req   models.UpdateScoreRequest
				exp   expected
			}{
				{
					auth:  models.AuthInfo{IsConnected: true, UserID: userA.Id},
					param: matchNoConsensus.Id,
					req:   models.NewUpdateScoreRequestFixture().WithScore1(1).WithScore2(0),
					exp:   expected{code: http.StatusOK, state: models.ManqueScore, score1: 1, score2: 0},
				},
				{
					auth:  models.AuthInfo{IsConnected: true, UserID: userB.Id},
					param: matchNoConsensus.Id,
					req:   models.NewUpdateScoreRequestFixture().WithScore1(2).WithScore2(2),
					exp:   expected{code: http.StatusOK, state: models.ManqueScore, score1: 2, score2: 2},
				},
			},
		},
		{
			name:     "Same team second vote blocked -> error and no ELO update",
			fixtures: fixtures,
			steps: []struct {
				auth  models.AuthInfo
				param string
				req   models.UpdateScoreRequest
				exp   expected
			}{
				{
					auth:  models.AuthInfo{IsConnected: true, UserID: userA.Id},
					param: matchTeamTwice.Id,
					req:   models.NewUpdateScoreRequestFixture().WithScore1(5).WithScore2(4),
					exp:   expected{code: http.StatusOK, state: models.ManqueScore, score1: 5, score2: 4},
				},
				{
					auth:  models.AuthInfo{IsConnected: true, UserID: userC.Id},
					param: matchTeamTwice.Id,
					req:   models.NewUpdateScoreRequestFixture().WithScore1(7).WithScore2(7),
					exp:   expected{code: http.StatusBadRequest, state: models.ManqueScore, score1: 5, score2: 4, errorContains: "team already has a vote"},
				},
			},
		},
	}

	const expectedDefaultElo = 1000
	const expectedDelta = 16

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()

			s.loadFixtures(tc.fixtures)

			for stepIdx, step := range tc.steps {
				body, err := json.Marshal(step.req)
				require.NoError(t, err)

				url := "/score/match/" + step.param
				r := httptest.NewRequest("PATCH", url, bytes.NewReader(body))
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("id", step.param)
				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

				w := httptest.NewRecorder()

				err = s.UpdateMatchScore(w, r, step.auth)
				require.NoError(t, err)

				resp := w.Result()
				defer func(Body io.ReadCloser) {
					_ = Body.Close()
				}(resp.Body)
				require.Equal(t, step.exp.code, resp.StatusCode)

				m, dbErr := s.db.GetMatchById(context.Background(), step.param)
				require.NoError(t, dbErr)
				require.NotNil(t, m)
				require.Equal(t, step.exp.state, m.CurrentState)
				require.NotNil(t, m.Score1)
				require.NotNil(t, m.Score2)
				require.Equal(t, step.exp.score1, *m.Score1)
				require.Equal(t, step.exp.score2, *m.Score2)

				if step.exp.errorContains != "" {
					b, _ := io.ReadAll(resp.Body)
					require.Contains(t, string(b), step.exp.errorContains)
				}

				ctx := context.Background()

				switch step.param {
				case matchConsensus.Id:
					if stepIdx == 0 {
						rA, err := s.db.GetRankingByUserAndCourt(ctx, userA.Id, court.Id)
						require.NoError(t, err)
						require.Nil(t, rA)

						rB, err := s.db.GetRankingByUserAndCourt(ctx, userB.Id, court.Id)
						require.NoError(t, err)
						require.Nil(t, rB)
					} else {
						rA, err := s.db.GetRankingByUserAndCourt(ctx, userA.Id, court.Id)
						require.NoError(t, err)
						require.NotNil(t, rA)
						require.Equal(t, expectedDefaultElo+expectedDelta, rA.Elo)

						rB, err := s.db.GetRankingByUserAndCourt(ctx, userB.Id, court.Id)
						require.NoError(t, err)
						require.NotNil(t, rB)
						require.Equal(t, expectedDefaultElo-expectedDelta, rB.Elo)
					}

				case matchNoConsensus.Id:
					rA, err := s.db.GetRankingByUserAndCourt(ctx, userA.Id, court.Id)
					require.NoError(t, err)
					require.Nil(t, rA)

					rB, err := s.db.GetRankingByUserAndCourt(ctx, userB.Id, court.Id)
					require.NoError(t, err)
					require.Nil(t, rB)

				case matchTeamTwice.Id:
					rA, err := s.db.GetRankingByUserAndCourt(ctx, userA.Id, court.Id)
					require.NoError(t, err)
					require.Nil(t, rA)

					rC, err := s.db.GetRankingByUserAndCourt(ctx, userC.Id, court.Id)
					require.NoError(t, err)
					require.Nil(t, rC)
				}
			}
		})
	}
}

func Test_JoinMatch(t *testing.T) {
	type expected struct {
		bodyJSON    string
		code        int
		errorMsg    string
		checkJoined bool
		wantState   *models.MatchState

		checkRanking bool
		wantElo      *int
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
	uTeam1A := models.NewDBUsersFixture().WithUsername("t1_a").WithEmail("t1_a@gmail.com")
	uTeam1B := models.NewDBUsersFixture().WithUsername("t1_b").WithEmail("t1_b@gmail.com")
	uTeam2A := models.NewDBUsersFixture().WithUsername("t2_a").WithEmail("t2_a@gmail.com")
	teammate := models.NewDBUsersFixture().
		WithUsername("teammate").
		WithEmail("other_email@gmail.com")

	match4 := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithParticipantNber(4).
		WithCurrentState(models.ManqueJoueur)

	defaultElo := 1000
	existingElo := 1337

	testCases := []testCase{
		{
			name: "User successfully joins match (no one else) -> still Manque joueur + creates default ranking",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user},
			},
			param: match.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			expected: expected{
				bodyJSON:     `{"team": 1}`,
				code:         http.StatusOK,
				checkJoined:  true,
				wantState:    ptr(models.ManqueJoueur),
				checkRanking: true,
				wantElo:      &defaultElo,
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
						WithMatchId(match.Id).
						WithTeam(1),
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
		{
			name: "User joins and fills the match -> match becomes Valide (ParticipantNber=4) + creates default ranking",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match4},
				Users:   []models.DBUsers{user, uTeam1A, uTeam1B, uTeam2A},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(uTeam1A.Id).
						WithMatchId(match4.Id).
						WithTeam(1),
					models.NewDBUserMatchFixture().
						WithUserId(uTeam1B.Id).
						WithMatchId(match4.Id).
						WithTeam(1),
					models.NewDBUserMatchFixture().
						WithUserId(uTeam2A.Id).
						WithMatchId(match4.Id).
						WithTeam(2),
				},
			},
			param: match4.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			expected: expected{
				bodyJSON:     `{"team": 2}`,
				code:         http.StatusOK,
				checkJoined:  true,
				wantState:    ptr(models.Valide),
				checkRanking: true,
				wantElo:      &defaultElo,
			},
		},
		{
			name: "Ranking already exists -> not overwritten",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user},
				Rankings: []models.DBRanking{
					{
						UserID:    user.Id,
						CourtID:   court.Id,
						Elo:       existingElo,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
			},
			param: match.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			expected: expected{
				bodyJSON:     `{"team": 1}`,
				code:         http.StatusOK,
				checkJoined:  true,
				wantState:    ptr(models.ManqueJoueur),
				checkRanking: true,
				wantElo:      &existingElo,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
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

			if c.expected.checkJoined || c.expected.wantState != nil || c.expected.checkRanking {
				ctx := context.Background()
				mid := c.param
				if mid == "" {
					mid = match.Id
				}

				updatedMatch, err := s.db.GetMatchById(ctx, mid)
				require.NoError(t, err)
				require.NotNil(t, updatedMatch)

				if c.expected.checkJoined {
					joined, err := s.db.IsUserInMatch(ctx, c.auth.UserID, mid)
					require.NoError(t, err)
					require.True(t, joined)
				}
				if c.expected.wantState != nil {
					require.Equal(t, *c.expected.wantState, updatedMatch.CurrentState)
				}
				if c.expected.checkRanking {
					rk, err := s.db.GetRankingByUserAndCourt(ctx, c.auth.UserID, updatedMatch.CourtID)
					require.NoError(t, err)
					require.NotNil(t, rk)
					if c.expected.wantElo != nil {
						require.Equal(t, *c.expected.wantElo, rk.Elo)
					}
				}
			}
		})
	}
}

func Test_StartMatch(t *testing.T) {
	type expected struct {
		code        int
		finalState  models.MatchState
		shouldCheck bool
	}

	type testCase struct {
		name     string
		auth     models.AuthInfo
		paramID  string
		fixtures DBFixtures
		expected expected
	}

	user := models.NewDBUsersFixture()
	court := models.NewDBCourtFixture()
	matchValide := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Valide)
	matchWrongState := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.ManqueJoueur)

	testCases := []testCase{
		{
			name:    "Happy path: start match",
			auth:    models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID: matchValide.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{matchValide},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(user.Id).WithMatchId(matchValide.Id),
				},
			},
			expected: expected{
				code:        http.StatusOK,
				finalState:  models.EnCours,
				shouldCheck: true,
			},
		},
		{
			name:     "Missing match ID",
			auth:     models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID:  "",
			fixtures: DBFixtures{},
			expected: expected{code: http.StatusBadRequest},
		},
		{
			name:     "Unauthorized",
			auth:     models.AuthInfo{IsConnected: false, UserID: user.Id},
			paramID:  matchValide.Id,
			fixtures: DBFixtures{},
			expected: expected{code: http.StatusUnauthorized},
		},
		{
			name:    "Match not found",
			auth:    models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID: uuid.NewString(),
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Users:  []models.DBUsers{user},
			},
			expected: expected{code: http.StatusNotFound},
		},
		{
			name:    "Wrong state",
			auth:    models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID: matchWrongState.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{matchWrongState},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(user.Id).WithMatchId(matchWrongState.Id),
				},
			},
			expected: expected{code: http.StatusBadRequest},
		},
		{
			name:    "User not in match",
			auth:    models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID: matchValide.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{matchValide},
				Users:   []models.DBUsers{user},
			},
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
			s.loadFixtures(c.fixtures)

			url := "/match/" + c.paramID + "/start"
			r := httptest.NewRequest("PATCH", url, nil)
			routeCtx := chi.NewRouteContext()
			if c.paramID != "" {
				routeCtx.URLParams.Add("id", c.paramID)
			}
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			err := s.StartMatch(w, r, c.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, c.expected.code, resp.StatusCode)

			if c.expected.shouldCheck {
				m, err := s.db.GetMatchById(context.Background(), c.paramID)
				require.NoError(t, err)
				require.NotNil(t, m)
				require.Equal(t, c.expected.finalState, m.CurrentState)
				require.WithinDuration(t, time.Now(), m.Date, time.Minute)
			}
		})
	}
}

func Test_FinishMatch(t *testing.T) {
	type expected struct {
		code        int
		finalState  models.MatchState
		shouldCheck bool
	}

	type testCase struct {
		name     string
		auth     models.AuthInfo
		paramID  string
		fixtures DBFixtures
		expected expected
	}

	user := models.NewDBUsersFixture()
	court := models.NewDBCourtFixture()
	matchEnCours := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.EnCours)
	matchWrongState := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Valide)

	tests := []testCase{
		{
			name:    "Happy path: finish match -> ManqueScore",
			auth:    models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID: matchEnCours.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{matchEnCours},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(user.Id).WithMatchId(matchEnCours.Id),
				},
			},
			expected: expected{
				code:        http.StatusOK,
				finalState:  models.ManqueScore,
				shouldCheck: true,
			},
		},
		{
			name:     "Missing match ID",
			auth:     models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID:  "",
			fixtures: DBFixtures{},
			expected: expected{code: http.StatusBadRequest},
		},
		{
			name:     "Unauthorized",
			auth:     models.AuthInfo{IsConnected: false, UserID: user.Id},
			paramID:  matchEnCours.Id,
			fixtures: DBFixtures{},
			expected: expected{code: http.StatusUnauthorized},
		},
		{
			name:    "Match not found",
			auth:    models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID: uuid.NewString(),
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Users:  []models.DBUsers{user},
			},
			expected: expected{code: http.StatusNotFound},
		},
		{
			name:    "Wrong state",
			auth:    models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID: matchWrongState.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{matchWrongState},
				Users:   []models.DBUsers{user},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(user.Id).WithMatchId(matchWrongState.Id),
				},
			},
			expected: expected{code: http.StatusBadRequest},
		},
		{
			name:    "User not in match",
			auth:    models.AuthInfo{IsConnected: true, UserID: user.Id},
			paramID: matchEnCours.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{matchEnCours},
				Users:   []models.DBUsers{user},
			},
			expected: expected{code: http.StatusBadRequest},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
			s.loadFixtures(tc.fixtures)

			url := "/match/" + tc.paramID + "/finish"
			r := httptest.NewRequest("PATCH", url, nil)
			routeCtx := chi.NewRouteContext()
			if tc.paramID != "" {
				routeCtx.URLParams.Add("id", tc.paramID)
			}
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			err := s.FinishMatch(w, r, tc.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, tc.expected.code, resp.StatusCode)

			if tc.expected.shouldCheck {
				m, err := s.db.GetMatchById(context.Background(), tc.paramID)
				require.NoError(t, err)
				require.NotNil(t, m)
				require.Equal(t, tc.expected.finalState, m.CurrentState)
			}
		})
	}
}
