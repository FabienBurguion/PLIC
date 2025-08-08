package main

import (
	"PLIC/models"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_GetUserById(t *testing.T) {
	type expected struct {
		code  int
		check func(t *testing.T, res models.UserResponse)
	}

	type testCase struct {
		name     string
		param    string
		fixtures DBFixtures
		expected expected
	}

	court1 := models.NewDBCourtFixture()
	court2 := models.NewDBCourtFixture()

	userWithData := models.NewDBUsersFixture()
	userNoMatch := models.NewDBUsersFixture()

	match1 := models.NewDBMatchesFixture().
		WithCourtId(court1.Id).
		WithCurrentState(models.Termine).
		WithSport(models.Foot)
	match2 := models.NewDBMatchesFixture().
		WithCourtId(court2.Id).
		WithCurrentState(models.Termine).
		WithSport(models.Basket)
	match3 := models.NewDBMatchesFixture().
		WithCourtId(court1.Id).
		WithCurrentState(models.Termine).
		WithSport(models.PingPong)

	testCases := []testCase{
		{
			name:  "User with full match history",
			param: userWithData.Id,
			fixtures: DBFixtures{
				Users:   []models.DBUsers{userWithData},
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2, match3},
				UserMatches: []models.DBUserMatch{
					{UserID: userWithData.Id, MatchID: match1.Id},
					{UserID: userWithData.Id, MatchID: match2.Id},
					{UserID: userWithData.Id, MatchID: match3.Id},
				},
			},
			expected: expected{
				code: http.StatusOK,
				check: func(t *testing.T, res models.UserResponse) {
					require.Equal(t, userWithData.Username, res.Username)
					require.Equal(t, userWithData.Bio, res.Bio)
					require.Equal(t, userWithData.CreatedAt.Unix(), res.CreatedAt.Unix())
					require.Equal(t, 3, res.NbMatches)
					require.Equal(t, 2, res.VisitedFields)

					if res.FavoriteSport != nil {
						require.Contains(t, []models.Sport{models.Foot, models.Basket, models.PingPong}, *res.FavoriteSport)
					} else {
						t.Log("FavoriteSport is nil (res if no match was inserted)")
					}
					require.ElementsMatch(t, []models.Sport{models.Foot, models.Basket, models.PingPong}, res.Sports)

					if res.FavoriteCity != nil {
						require.Contains(t, []string{"Paris", "Lyon"}, *res.FavoriteCity)
					} else {
						t.Log("FavoriteCity is nil (res if no match was inserted)")
					}

					require.NotNil(t, res.FavoriteField)

					require.Len(t, res.Fields, 0)
				},
			},
		},
		{
			name:  "User with no matches",
			param: userNoMatch.Id,
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					userNoMatch,
				},
			},
			expected: expected{
				code: http.StatusOK,
				check: func(t *testing.T, res models.UserResponse) {
					require.Equal(t, 0, res.NbMatches)
					require.Equal(t, 0, res.VisitedFields)
					require.Nil(t, res.FavoriteSport)
					require.Nil(t, res.FavoriteCity)
					require.Nil(t, res.FavoriteField)
					require.Empty(t, res.Sports)
					require.Empty(t, res.Fields)
				},
			},
		},
		{
			name:     "User not found",
			param:    uuid.NewString(),
			fixtures: DBFixtures{},
			expected: expected{
				code:  http.StatusNotFound,
				check: nil,
			},
		},
		{
			name:     "Missing ID",
			param:    "",
			fixtures: DBFixtures{},
			expected: expected{
				code:  http.StatusBadRequest,
				check: nil,
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
			s.loadFixtures(c.fixtures)

			url := "/users"
			if c.param != "" {
				url += "/" + c.param
			}

			r := httptest.NewRequest("GET", url, nil)
			r.Header.Set("Content-Type", "application/json")
			if c.param != "" {
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("id", c.param)
				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
			}

			w := httptest.NewRecorder()

			err := s.GetUserById(w, r, models.AuthInfo{})
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			require.Equal(t, c.expected.code, resp.StatusCode)

			if c.expected.code == http.StatusOK && c.expected.check != nil {
				bodyBytes, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var res models.UserResponse
				err = json.Unmarshal(bodyBytes, &res)
				require.NoError(t, err)

				c.expected.check(t, res)
			}
		})
	}
}

func Test_PatchUser(t *testing.T) {
	type expected struct {
		res  *models.DBUsers
		code int
	}

	type testCase struct {
		name      string
		fixtures  DBFixtures
		param     models.UserPatchRequest
		auth      models.AuthInfo
		urlUserId string
		expected  expected
	}

	userId := uuid.NewString()
	originalUsername := "Fabien"
	originalEmail := "<EMAIL>"
	originalBio := "A bio"

	newUsername := "New username"
	newEmail := "<EMAIL2>"
	newBio := "A new bio"
	newCurrentFieldId := "a-new-current-field-id"

	testCases := []testCase{
		{
			name: "Update all fields",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithUsername(originalUsername).
						WithEmail(originalEmail).
						WithBio(originalBio),
				},
			},
			param: models.UserPatchRequest{
				Username: ptr(newUsername),
				Email:    ptr(newEmail),
				Bio:      ptr(newBio),
			},
			auth: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			urlUserId: userId,
			expected: expected{
				code: 200,
				res: &models.DBUsers{
					Id:       userId,
					Username: newUsername,
					Email:    newEmail,
					Bio:      ptr(newBio),
				},
			},
		},
		{
			name: "User not found",
			fixtures: DBFixtures{
				Users: []models.DBUsers{},
			},
			param: models.UserPatchRequest{
				Bio: ptr(newBio),
			},
			auth: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			urlUserId: userId,
			expected: expected{
				code: 200,
				res:  nil,
			},
		},
		{
			name: "Not connected",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().WithId(userId),
				},
			},
			param: models.UserPatchRequest{
				Bio: ptr(newBio),
			},
			auth: models.AuthInfo{
				IsConnected: false,
				UserID:      userId,
			},
			urlUserId: userId,
			expected: expected{
				code: 403,
				res:  nil,
			},
		},
		{
			name: "Wrong param ID",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().WithId(userId),
				},
			},
			param: models.UserPatchRequest{
				Bio: ptr(newBio),
			},
			auth: models.AuthInfo{
				IsConnected: true,
				UserID:      "another-param-id",
			},
			urlUserId: userId,
			expected: expected{
				code: 403,
				res:  nil,
			},
		},
		{
			name: "Current field id",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().WithId(userId),
				},
			},
			param: models.UserPatchRequest{
				CurrentFieldId: ptr(newCurrentFieldId),
			},
			auth: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			urlUserId: userId,
			expected: expected{
				code: 200,
				res: &models.DBUsers{
					Id:             userId,
					Username:       "username",
					Email:          "an email",
					Bio:            ptr("a bio"),
					CurrentFieldId: &newCurrentFieldId,
				},
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
			s.loadFixtures(c.fixtures)

			body, err := json.Marshal(c.param)
			require.NoError(t, err)

			r := httptest.NewRequest("PATCH", "/users/"+c.urlUserId, bytes.NewReader(body))
			r.Header.Set("Content-Type", "application/json")

			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", c.urlUserId)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()

			err = s.PatchUser(w, r, c.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			require.Equal(t, c.expected.code, resp.StatusCode)

			if c.expected.res != nil {
				updated, err := s.db.GetUserById(r.Context(), c.urlUserId)
				require.NoError(t, err)

				require.Equal(t, c.expected.res.Username, updated.Username)
				require.Equal(t, c.expected.res.Email, updated.Email)
				require.Equal(t, *c.expected.res.Bio, *updated.Bio)
				if c.expected.res.CurrentFieldId != nil {
					require.Equal(t, *c.expected.res.CurrentFieldId, *updated.CurrentFieldId)
				}
			}
		})
	}
}

func Test_DeleteUser(t *testing.T) {
	type testCase struct {
		name     string
		param    string
		auth     models.AuthInfo
		fixtures DBFixtures
		expected int
	}

	userId := uuid.NewString()

	testCases := []testCase{
		{
			name: "Basic test",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId),
				},
			},
			param: userId,
			auth: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			expected: 200,
		},
		{
			name: "No auth info",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId),
				},
			},
			param: userId,
			auth: models.AuthInfo{
				IsConnected: false,
				UserID:      "",
			},
			expected: 403,
		},
		{
			name: "No userId",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId),
				},
			},
			param: "",
			auth: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			expected: 400,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			t.Parallel()

			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(c.fixtures)

			r := httptest.NewRequest("PATCH", "/users/"+c.param, nil)
			r.Header.Set("Content-Type", "application/json")

			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", c.param)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()

			err := s.DeleteUser(w, r, c.auth)
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			require.Equal(t, c.expected, resp.StatusCode)

			if c.expected != 200 {
				return
			}

			user, err := s.db.GetUserById(ctx, c.param)
			require.NoError(t, err)
			require.Nil(t, user)
		})
	}
}
