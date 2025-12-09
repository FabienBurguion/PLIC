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

	user := models.NewDBUsersFixture()

	match := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCreatorId(user.Id)

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
	user := models.NewDBUsersFixture()
	match := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCreatorId(user.Id)

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

	user := models.NewDBUsersFixture()

	match1 := models.NewDBMatchesFixture().
		WithCourtId(court1.Id).
		WithSport(models.Foot).
		WithCurrentState(models.Termine).
		WithCreatorId(user.Id)
	match2 := models.NewDBMatchesFixture().
		WithCourtId(court2.Id).
		WithSport(models.Basket).
		WithCurrentState(models.Termine).
		WithCreatorId(user.Id)

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
				Users:   []models.DBUsers{user},
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
				Users:   []models.DBUsers{user},
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

	user := models.NewDBUsersFixture()

	match1 := models.NewDBMatchesFixture().
		WithCourtId(court1.Id).
		WithCreatorId(user.Id)
	match2 := models.NewDBMatchesFixture().
		WithCourtId(court2.Id).
		WithCreatorId(user.Id)

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

	// On fixe explicitement le sport utilisé dans ce test
	sport := models.Basket

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
				WithCourtId(court.Id).
				WithSport(sport),
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
				WithCourtId(court.Id).
				WithSport(sport),
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
				WithCourtId(court.Id).
				WithSport(sport),
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
					models.NewDBRankingFixture().
						WithUserId(user.Id).
						WithCourtId(court.Id).
						WithSport(sport).
						WithElo(existingElo),
				},
			},
			insertCourt: false,
			param: models.NewMatchRequestFixture().
				WithCourtId(court.Id).
				WithSport(sport),
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
				rk, err := s.db.GetRankingByUserCourtSport(ctx, user.Id, c.param.CourtID, c.param.Sport)
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
	const sport = models.Foot

	userA := models.NewDBUsersFixture()
	userB := models.NewDBUsersFixture().WithUsername("userB").WithEmail("emailB@example.com")
	userC := models.NewDBUsersFixture().WithUsername("userC").WithEmail("emailC@example.com")

	matchConsensus := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithSport(sport).
		WithCurrentState(models.ManqueScore).
		WithCreatorId(userA.Id)

	matchNoConsensus := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithSport(sport).
		WithCurrentState(models.ManqueScore).
		WithCreatorId(userA.Id)

	matchTeamTwice := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithSport(sport).
		WithCurrentState(models.ManqueScore).
		WithCreatorId(userA.Id)

	baseFixtures := DBFixtures{
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

	defaultElo := 1000
	withRankings := baseFixtures
	withRankings.Rankings = []models.DBRanking{
		models.NewDBRankingFixture().WithUserId(userA.Id).WithCourtId(court.Id).WithSport(sport).WithElo(defaultElo),
		models.NewDBRankingFixture().WithUserId(userB.Id).WithCourtId(court.Id).WithSport(sport).WithElo(defaultElo),
		models.NewDBRankingFixture().WithUserId(userA.Id).WithCourtId(court.Id).WithSport(sport).WithElo(defaultElo),
		models.NewDBRankingFixture().WithUserId(userC.Id).WithCourtId(court.Id).WithSport(sport).WithElo(defaultElo),
	}

	testCases := []testCase{
		{
			name:     "Consensus: two teams same score -> match becomes Termine and ELO updates",
			fixtures: withRankings,
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
			name:     "No consensus: second team different score -> state unchanged, scores are last proposal, no ELO update",
			fixtures: baseFixtures,
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
			fixtures: baseFixtures,
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

	const expectedDelta = 16

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()

			mockMailer := mailer.NewMockMailer()
			s.mailer = mockMailer

			s.loadFixtures(tc.fixtures)

			sentResultSoFar := 0

			for stepIdx, step := range tc.steps {
				body, err := json.Marshal(step.req)
				require.NoError(t, err)

				url := "/match/" + step.param + "/score"
				r := httptest.NewRequest("PATCH", url, bytes.NewReader(body))
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("id", step.param)
				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

				w := httptest.NewRecorder()

				err = s.UpdateMatchScore(w, r, step.auth)
				require.NoError(t, err)

				resp := w.Result()
				defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
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

				switch step.param {
				case matchConsensus.Id:
					if stepIdx == 1 {
						ctx := context.Background()
						ums, err := s.db.GetUserMatchesByMatchID(ctx, step.param)
						require.NoError(t, err)
						sentResultSoFar += len(ums)
					}
				}
				require.Equal(t, sentResultSoFar, mockMailer.GetSentCounts("result"),
					"unexpected number of result emails after step %d of %q", stepIdx, tc.name)

				ctx := context.Background()
				switch step.param {
				case matchConsensus.Id:
					if stepIdx == 0 {
						rA, err := s.db.GetRankingByUserCourtSport(ctx, userA.Id, court.Id, sport)
						require.NoError(t, err)
						require.NotNil(t, rA)
						require.Equal(t, defaultElo, rA.Elo)

						rB, err := s.db.GetRankingByUserCourtSport(ctx, userB.Id, court.Id, sport)
						require.NoError(t, err)
						require.NotNil(t, rB)
						require.Equal(t, defaultElo, rB.Elo)
					} else {
						rA, err := s.db.GetRankingByUserCourtSport(ctx, userA.Id, court.Id, sport)
						require.NoError(t, err)
						require.NotNil(t, rA)
						require.Equal(t, defaultElo+expectedDelta, rA.Elo)

						rB, err := s.db.GetRankingByUserCourtSport(ctx, userB.Id, court.Id, sport)
						require.NoError(t, err)
						require.NotNil(t, rB)
						require.Equal(t, defaultElo-expectedDelta, rB.Elo)
					}

				case matchNoConsensus.Id:

				case matchTeamTwice.Id:
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

	const sport = models.Basket
	user := models.NewDBUsersFixture()

	match := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithSport(sport).
		WithParticipantNber(2).
		WithCurrentState(models.ManqueJoueur).
		WithCreatorId(user.Id)

	uTeam1A := models.NewDBUsersFixture().WithUsername("t1_a").WithEmail("t1_a@gmail.com")
	uTeam1B := models.NewDBUsersFixture().WithUsername("t1_b").WithEmail("t1_b@gmail.com")
	uTeam2A := models.NewDBUsersFixture().WithUsername("t2_a").WithEmail("t2_a@gmail.com")
	teammate := models.NewDBUsersFixture().
		WithUsername("teammate").
		WithEmail("other_email@gmail.com")

	match4 := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithSport(sport).
		WithParticipantNber(4).
		WithCurrentState(models.ManqueJoueur).
		WithCreatorId(user.Id)

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
					models.NewDBRankingFixture().
						WithUserId(user.Id).
						WithCourtId(court.Id).
						WithSport(sport).
						WithElo(existingElo),
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
			defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

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
					rk, err := s.db.GetRankingByUserCourtSport(ctx, c.auth.UserID, updatedMatch.CourtID, sport)
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
		WithCurrentState(models.Valide).
		WithCreatorId(user.Id)
	matchWrongState := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.ManqueJoueur).
		WithCreatorId(user.Id)

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
		WithCurrentState(models.EnCours).
		WithCreatorId(user.Id)
	matchWrongState := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Valide).
		WithCreatorId(user.Id)

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

func Test_GetMatchVoteStatus(t *testing.T) {
	type expected struct {
		code        int
		errContains string
		wantTeam    *int
		myHasVoted  *bool
		opHasVoted  *bool
		myScore     *models.ScorePair
		opScore     *models.ScorePair
	}

	type testCase struct {
		name     string
		auth     models.AuthInfo
		paramID  string
		fixtures DBFixtures
		prepare  func(t *testing.T, s *Service, matchID, userID string)
		expected expected
	}

	court := models.NewDBCourtFixture()

	userA := models.NewDBUsersFixture().WithUsername("userA").WithEmail("a@example.com")
	userB := models.NewDBUsersFixture().WithUsername("userB").WithEmail("b@example.com")
	userC := models.NewDBUsersFixture().WithUsername("userC").WithEmail("c@example.com")

	matchTermine := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.ManqueScore).
		WithCreatorId(userA.Id)

	matchTermineOnlyOpp := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.ManqueScore).
		WithCreatorId(userA.Id)

	makeBool := func(b bool) *bool { return &b }
	makeTeam := func(t int) *int { return &t }
	makeScore := func(s1, s2 int) *models.ScorePair { return &models.ScorePair{Score1: s1, Score2: s2} }

	tests := []testCase{
		{
			name:    "Happy path: both teams voted, user in team 1",
			auth:    models.AuthInfo{IsConnected: true, UserID: userA.Id},
			paramID: matchTermine.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{matchTermine},
				Users:   []models.DBUsers{userA, userB},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(userA.Id).WithMatchId(matchTermine.Id).WithTeam(1),
					models.NewDBUserMatchFixture().WithUserId(userB.Id).WithMatchId(matchTermine.Id).WithTeam(2),
				},
			},
			prepare: func(t *testing.T, s *Service, matchID, _ string) {
				ctx := context.Background()
				require.NoError(t, s.db.UpsertMatchScoreVote(ctx, models.DBMatchScoreVote{
					MatchId: matchID, UserId: userA.Id, Team: 1, Score1: 3, Score2: 2,
				}))
				require.NoError(t, s.db.UpsertMatchScoreVote(ctx, models.DBMatchScoreVote{
					MatchId: matchID, UserId: userB.Id, Team: 2, Score1: 1, Score2: 4,
				}))
			},
			expected: expected{
				code:       http.StatusOK,
				wantTeam:   makeTeam(1),
				myHasVoted: makeBool(true),
				opHasVoted: makeBool(true),
				myScore:    makeScore(3, 2),
				opScore:    makeScore(1, 4),
			},
		},
		{
			name:    "Only opponent voted: my team has no vote",
			auth:    models.AuthInfo{IsConnected: true, UserID: userA.Id},
			paramID: matchTermineOnlyOpp.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{matchTermineOnlyOpp},
				Users:   []models.DBUsers{userA, userB},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(userA.Id).WithMatchId(matchTermineOnlyOpp.Id).WithTeam(1),
					models.NewDBUserMatchFixture().WithUserId(userB.Id).WithMatchId(matchTermineOnlyOpp.Id).WithTeam(2),
				},
			},
			prepare: func(t *testing.T, s *Service, matchID, _ string) {
				ctx := context.Background()
				require.NoError(t, s.db.UpsertMatchScoreVote(ctx, models.DBMatchScoreVote{
					MatchId: matchID, UserId: userB.Id, Team: 2, Score1: 9, Score2: 9,
				}))
			},
			expected: expected{
				code:       http.StatusOK,
				wantTeam:   makeTeam(1),
				myHasVoted: makeBool(false),
				opHasVoted: makeBool(true),
				myScore:    nil,
				opScore:    makeScore(9, 9),
			},
		},
		{
			name:    "User not in match -> 404",
			auth:    models.AuthInfo{IsConnected: true, UserID: userC.Id},
			paramID: matchTermine.Id,
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{matchTermine},
				Users:   []models.DBUsers{userA, userB, userC},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(userA.Id).WithMatchId(matchTermine.Id).WithTeam(1),
					models.NewDBUserMatchFixture().WithUserId(userB.Id).WithMatchId(matchTermine.Id).WithTeam(2),
				},
			},
			expected: expected{
				code:        http.StatusNotFound,
				errContains: "user in this match not found",
			},
		},
		{
			name:     "Unauthorized",
			auth:     models.AuthInfo{IsConnected: false, UserID: userA.Id},
			paramID:  matchTermine.Id,
			fixtures: DBFixtures{},
			expected: expected{
				code:        http.StatusUnauthorized,
				errContains: "not authorized",
			},
		},
		{
			name:     "Missing match ID",
			auth:     models.AuthInfo{IsConnected: true, UserID: userA.Id},
			paramID:  "",
			fixtures: DBFixtures{},
			expected: expected{
				code:        http.StatusBadRequest,
				errContains: "missing match ID",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()

			s.loadFixtures(tc.fixtures)

			if tc.prepare != nil && tc.paramID != "" && tc.auth.UserID != "" {
				tc.prepare(t, s, tc.paramID, tc.auth.UserID)
			}

			url := "/match/" + tc.paramID + "/vote-status"
			r := httptest.NewRequest("GET", url, nil)
			routeCtx := chi.NewRouteContext()
			if tc.paramID != "" {
				routeCtx.URLParams.Add("id", tc.paramID)
			}
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			err := s.GetMatchVoteStatus(w, r, tc.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
			require.Equal(t, tc.expected.code, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)

			if tc.expected.code != http.StatusOK {
				if tc.expected.errContains != "" {
					require.Contains(t, string(body), tc.expected.errContains)
				}
				return
			}

			var res models.MatchVoteStatusResponse
			require.NoError(t, json.Unmarshal(body, &res))

			if tc.expected.wantTeam != nil {
				require.Equal(t, *tc.expected.wantTeam, res.PlayerTeam)
			}
			if tc.expected.myHasVoted != nil {
				require.Equal(t, *tc.expected.myHasVoted, res.MyTeam.HasVoted)
			}
			if tc.expected.opHasVoted != nil {
				require.Equal(t, *tc.expected.opHasVoted, res.Opponent.HasVoted)
			}
			if tc.expected.myScore != nil {
				require.NotNil(t, res.MyTeam.Score)
				require.Equal(t, tc.expected.myScore.Score1, res.MyTeam.Score.Score1)
				require.Equal(t, tc.expected.myScore.Score2, res.MyTeam.Score.Score2)
			} else {
				require.Nil(t, res.MyTeam.Score)
			}
			if tc.expected.opScore != nil {
				require.NotNil(t, res.Opponent.Score)
				require.Equal(t, tc.expected.opScore.Score1, res.Opponent.Score.Score1)
				require.Equal(t, tc.expected.opScore.Score2, res.Opponent.Score.Score2)
			} else {
				require.Nil(t, res.Opponent.Score)
			}
		})
	}
}
