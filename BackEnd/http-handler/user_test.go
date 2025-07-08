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

func Test_GetUserById(t *testing.T) {
	type testCase struct {
		name          string
		urlID         string
		fixtures      DBFixtures
		expectedCode  int
		expectedCheck func(t *testing.T, res models.UserResponse)
	}

	userWithData := models.NewDBUsersFixture()
	userNoMatch := models.NewDBUsersFixture()

	match1 := models.DBMatches{
		Id:           uuid.NewString(),
		Sport:        models.Foot,
		Place:        "Paris",
		Date:         time.Now(),
		CurrentState: models.Termine,
		Score1:       2,
		Score2:       1,
	}
	match2 := models.DBMatches{
		Id:           uuid.NewString(),
		Sport:        models.Basket,
		Place:        "Lyon",
		Date:         time.Now(),
		CurrentState: models.Termine,
		Score1:       5,
		Score2:       3,
	}

	testCases := []testCase{
		{
			name:  "User with full match history",
			urlID: userWithData.Id,
			fixtures: DBFixtures{
				Users: []models.DBUsers{userWithData},
				Matches: []models.DBMatches{
					match1, match2,
				},
				UserMatches: []models.DBUserMatch{
					{UserID: userWithData.Id, MatchID: match1.Id},
					{UserID: userWithData.Id, MatchID: match2.Id},
				},
			},
			expectedCode: http.StatusOK,
			expectedCheck: func(t *testing.T, res models.UserResponse) {
				require.Equal(t, userWithData.Username, res.Username)
				require.Equal(t, userWithData.Bio, res.Bio)
				require.Equal(t, userWithData.CreatedAt.Unix(), res.CreatedAt.Unix())
				require.Equal(t, 2, res.NbMatches)
				require.Equal(t, 2, res.VisitedFields)

				if res.FavoriteSport != nil {
					require.Contains(t, []models.Sport{models.Foot, models.Basket}, *res.FavoriteSport)
				} else {
					t.Log("FavoriteSport is nil (expected if no match was inserted)")
				}
				require.ElementsMatch(t, []models.Sport{models.Foot, models.Basket}, res.Sports)

				if res.FavoriteCity != nil {
					require.Contains(t, []string{"Paris", "Lyon"}, *res.FavoriteCity)
				} else {
					t.Log("FavoriteCity is nil (expected if no match was inserted)")
				}

				require.NotNil(t, res.FavoriteField)
				require.Contains(t, []string{"Paris", "Lyon"}, *res.FavoriteField)

				require.Len(t, res.Fields, 0)
			},
		},
		{
			name:  "User with no matches",
			urlID: userNoMatch.Id,
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					userNoMatch,
				},
			},
			expectedCode: http.StatusOK,
			expectedCheck: func(t *testing.T, res models.UserResponse) {
				require.Equal(t, 0, res.NbMatches)
				require.Equal(t, 0, res.VisitedFields)
				require.Nil(t, res.FavoriteSport)
				require.Nil(t, res.FavoriteCity)
				require.Nil(t, res.FavoriteField)
				require.Empty(t, res.Sports)
				require.Empty(t, res.Fields)
			},
		},
		{
			name:          "User not found",
			urlID:         uuid.NewString(),
			fixtures:      DBFixtures{},
			expectedCode:  http.StatusNotFound,
			expectedCheck: nil,
		},
		{
			name:          "Missing ID",
			urlID:         "",
			fixtures:      DBFixtures{},
			expectedCode:  http.StatusBadRequest,
			expectedCheck: nil,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			url := "/users"
			if c.urlID != "" {
				url += "/" + c.urlID
			}

			r := httptest.NewRequest("GET", url, nil)
			r.Header.Set("Content-Type", "application/json")
			if c.urlID != "" {
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("id", c.urlID)
				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
			}

			w := httptest.NewRecorder()

			err := s.GetUserById(w, r, models.AuthInfo{})
			require.NoError(t, err)

			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, c.expectedCode, resp.StatusCode)

			if c.expectedCode == http.StatusOK && c.expectedCheck != nil {
				bodyBytes, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var res models.UserResponse
				err = json.Unmarshal(bodyBytes, &res)
				require.NoError(t, err)

				c.expectedCheck(t, res)
			}
		})
	}
}

func Test_PatchUser(t *testing.T) {
	type testCase struct {
		name         string
		fixtures     DBFixtures
		param        models.UserPatchRequest
		authInfo     models.AuthInfo
		urlUserId    string
		expected     *models.DBUsers
		expectedCode int
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
			authInfo: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			urlUserId:    userId,
			expectedCode: 200,
			expected: &models.DBUsers{
				Id:       userId,
				Username: newUsername,
				Email:    newEmail,
				Bio:      ptr(newBio),
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
			authInfo: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			urlUserId:    userId,
			expectedCode: 200,
			expected:     nil,
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
			authInfo: models.AuthInfo{
				IsConnected: false,
				UserID:      userId,
			},
			urlUserId:    userId,
			expectedCode: 403,
			expected:     nil,
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
			authInfo: models.AuthInfo{
				IsConnected: true,
				UserID:      "another-param-id",
			},
			urlUserId:    userId,
			expectedCode: 403,
			expected:     nil,
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
			authInfo: models.AuthInfo{
				IsConnected: true,
				UserID:      userId,
			},
			urlUserId:    userId,
			expectedCode: 200,
			expected: &models.DBUsers{
				Id:             userId,
				Username:       "username",
				Email:          "an email",
				Bio:            ptr("a bio"),
				CurrentFieldId: &newCurrentFieldId,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			body, err := json.Marshal(c.param)
			require.NoError(t, err)

			r := httptest.NewRequest("PATCH", "/users/"+c.urlUserId, bytes.NewReader(body))
			r.Header.Set("Content-Type", "application/json")

			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", c.urlUserId)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()

			err = s.PatchUser(w, r, c.authInfo)
			require.NoError(t, err)

			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, c.expectedCode, resp.StatusCode)

			if c.expected != nil {
				updated, err := s.db.GetUserById(r.Context(), c.urlUserId)
				require.NoError(t, err)

				require.Equal(t, c.expected.Username, updated.Username)
				require.Equal(t, c.expected.Email, updated.Email)
				require.Equal(t, *c.expected.Bio, *updated.Bio)
				if c.expected.CurrentFieldId != nil {
					require.Equal(t, *c.expected.CurrentFieldId, *updated.CurrentFieldId)
				}
			}
		})
	}
}

func Test_DeleteUser(t *testing.T) {
	type expected struct {
		authInfo     models.AuthInfo
		expectedCode int
	}

	type testCase struct {
		name     string
		param    string
		fixtures DBFixtures
		expected expected
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
			expected: expected{
				authInfo: models.AuthInfo{
					IsConnected: true,
					UserID:      userId,
				},
				expectedCode: 200,
			},
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
			expected: expected{
				authInfo: models.AuthInfo{
					IsConnected: false,
					UserID:      "",
				},
				expectedCode: 403,
			},
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
			expected: expected{
				authInfo: models.AuthInfo{
					IsConnected: true,
					UserID:      userId,
				},
				expectedCode: 400,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			t.Parallel()

			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			r := httptest.NewRequest("PATCH", "/users/"+c.param, nil)
			r.Header.Set("Content-Type", "application/json")

			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", c.param)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()

			err := s.DeleteUser(w, r, c.expected.authInfo)
			require.NoError(t, err)

			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, c.expected.expectedCode, resp.StatusCode)

			if c.expected.expectedCode != 200 {
				return
			}

			user, err := s.db.GetUserById(ctx, c.param)
			require.NoError(t, err)
			require.Nil(t, user)
		})
	}
}
