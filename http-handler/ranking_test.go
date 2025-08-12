package main

import (
	"PLIC/models"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func Test_GetRankingByCourtId(t *testing.T) {
	type expected struct {
		code          int
		expectJSONArr bool
		expectErrMsg  string
		wantLen       int
		checkSorted   bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		auth     models.AuthInfo
		expected expected
	}

	court := models.NewDBCourtFixture()

	u1 := models.NewDBUsersFixture().WithUsername("alice").WithEmail("alice@example.com")
	u2 := models.NewDBUsersFixture().WithUsername("bob").WithEmail("bob@example.com")
	u3 := models.NewDBUsersFixture().WithUsername("carol").WithEmail("carol@example.com")

	r1 := models.NewDBRankingFixture().WithUserId(u1.Id).WithCourtId(court.Id).WithElo(1300)
	r2 := models.NewDBRankingFixture().WithUserId(u2.Id).WithCourtId(court.Id).WithElo(1100)
	r3 := models.NewDBRankingFixture().WithUserId(u3.Id).WithCourtId(court.Id).WithElo(1100)
	now := time.Now()
	for _, r := range []*models.DBRanking{&r1, &r2, &r3} {
		r.CreatedAt = now
		r.UpdatedAt = now
	}

	otherCourt := models.NewDBCourtFixture()
	u4 := models.NewDBUsersFixture().WithUsername("dave").WithEmail("dave@example.com")
	rOther := models.NewDBRankingFixture().WithUserId(u4.Id).WithCourtId(otherCourt.Id).WithElo(1500)
	rOther.CreatedAt = now
	rOther.UpdatedAt = now

	testCases := []testCase{
		{
			name: "OK - returns sorted list by elo then user_id",
			fixtures: DBFixtures{
				Courts:   []models.DBCourt{court},
				Users:    []models.DBUsers{u1, u2, u3},
				Rankings: []models.DBRanking{r1, r2, r3},
			},
			param: court.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: u1.Id},
			expected: expected{
				code:          http.StatusOK,
				expectJSONArr: true,
				wantLen:       3,
				checkSorted:   true,
			},
		},
		{
			name: "OK - empty list when no rankings",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Users:  []models.DBUsers{u1},
			},
			param: court.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: u1.Id},
			expected: expected{
				code:          http.StatusOK,
				expectJSONArr: true,
				wantLen:       0,
			},
		},
		{
			name: "OK - filters by courtID when multiple courts exist",
			fixtures: DBFixtures{
				Courts:   []models.DBCourt{court, otherCourt},
				Users:    []models.DBUsers{u1, u2, u3, u4},
				Rankings: []models.DBRanking{r1, r2, r3, rOther},
			},
			param: otherCourt.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: u4.Id},
			expected: expected{
				code:          http.StatusOK,
				expectJSONArr: true,
				wantLen:       1,
				checkSorted:   true,
			},
		},
		{
			name: "Bad request - missing id",
			fixtures: DBFixtures{
				Users: []models.DBUsers{u1},
			},
			param: "",
			auth:  models.AuthInfo{IsConnected: true, UserID: u1.Id},
			expected: expected{
				code:         http.StatusBadRequest,
				expectErrMsg: "missing court ID",
			},
		},
		{
			name: "Unauthorized",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Users:  []models.DBUsers{u1},
			},
			param: court.Id,
			auth:  models.AuthInfo{IsConnected: false, UserID: u1.Id},
			expected: expected{
				code:         http.StatusUnauthorized,
				expectErrMsg: "not authorized",
			},
		},
	}

	isSorted := func(arr []models.CourtRankingResponse) bool {
		for i := 1; i < len(arr); i++ {
			if arr[i-1].Elo > arr[i].Elo {
				return false
			}
			if arr[i-1].Elo == arr[i].Elo && arr[i-1].UserID > arr[i].UserID {
				return false
			}
		}
		return true
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()

			s.loadFixtures(tc.fixtures)

			url := "/ranking/court/" + tc.param
			r := httptest.NewRequest("GET", url, nil)
			routeCtx := chi.NewRouteContext()
			if tc.param != "" {
				routeCtx.URLParams.Add("id", tc.param)
			}
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			err := s.GetRankingByCourtId(w, r, tc.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

			require.Equal(t, tc.expected.code, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)

			if tc.expected.expectErrMsg != "" {
				require.Contains(t, string(body), tc.expected.expectErrMsg)
				return
			}

			if tc.expected.expectJSONArr {
				var out []models.CourtRankingResponse
				require.NoError(t, json.Unmarshal(body, &out))
				require.Len(t, out, tc.expected.wantLen)
				if tc.expected.checkSorted {
					require.True(t, isSorted(out), "ranking list should be sorted by (elo asc, user_id asc)")
				}
			}
		})
	}
}

func Test_GetUserFields(t *testing.T) {
	type expected struct {
		code          int
		expectBody    bool
		expectLen     int
		errorContains string
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		auth     models.AuthInfo
		expected expected
	}

	user := models.NewDBUsersFixture()
	c1 := models.NewDBCourtFixture().WithName("Court A")
	c2 := models.NewDBCourtFixture().WithName("Court B")

	rk1 := models.NewDBRankingFixture().WithUserId(user.Id).WithCourtId(c1.Id).WithElo(1200)
	rk2 := models.NewDBRankingFixture().WithUserId(user.Id).WithCourtId(c2.Id).WithElo(1300)

	tests := []testCase{
		{
			name: "success - returns ranked fields for user",
			fixtures: DBFixtures{
				Users:    []models.DBUsers{user},
				Courts:   []models.DBCourt{c1, c2},
				Rankings: []models.DBRanking{rk1, rk2},
			},
			param: user.Id,
			auth:  models.AuthInfo{IsConnected: true, UserID: user.Id},
			expected: expected{
				code:       http.StatusOK,
				expectBody: true,
				expectLen:  2,
			},
		},
		{
			name:     "missing userId",
			param:    "",
			auth:     models.AuthInfo{IsConnected: true, UserID: user.Id},
			expected: expected{code: http.StatusBadRequest, errorContains: "missing userId"},
		},
		{
			name:     "unauthorized",
			param:    user.Id,
			auth:     models.AuthInfo{IsConnected: false, UserID: user.Id},
			expected: expected{code: http.StatusUnauthorized, errorContains: "not authorized"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
			s.loadFixtures(tc.fixtures)

			url := "/user/" + tc.param + "/fields"
			r := httptest.NewRequest("GET", url, nil)

			routeCtx := chi.NewRouteContext()
			if tc.param != "" {
				routeCtx.URLParams.Add("userId", tc.param)
			}
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			err := s.GetUserFields(w, r, tc.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			require.Equal(t, tc.expected.code, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)

			if tc.expected.errorContains != "" {
				require.Contains(t, string(body), tc.expected.errorContains)
				return
			}

			if tc.expected.expectBody {
				var got []models.Field
				require.NoError(t, json.Unmarshal(body, &got))
				require.Len(t, got, tc.expected.expectLen)
			}
		})
	}
}
