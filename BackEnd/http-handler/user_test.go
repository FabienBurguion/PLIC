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
	"net/http/httptest"
	"testing"
)

func Test_GetUserById(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
		expected models.UserResponse
	}

	user := models.NewDBUsersFixture()

	testCases := []testCase{
		{
			name: "Basic test",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					user,
				},
			},
			expected: models.UserResponse{
				Username:       user.Username,
				ProfilePicture: ptr(user.Id + ".png"),
				Bio:            user.Bio,
				CreatedAt:      user.CreatedAt,
				VisitedFields:  0,
				Winrate:        ptr(100),
				FavoriteCity:   ptr("a wonderful city"),
				FavoriteSport:  ptr(models.Foot),
				FavoriteField:  ptr("a wonderful field"),
				Sports: []models.Sport{
					models.Basket,
					models.Foot,
				},
				Fields: []models.Field{
					{
						Ranking: 1,
						Name:    "a wonderful field",
						Score:   1000,
					},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			r := httptest.NewRequest("GET", "/users/"+user.Id, nil)
			r.Header.Set("Content-Type", "application/json")
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", user.Id)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
			w := httptest.NewRecorder()

			err := s.GetUserById(w, r, models.AuthInfo{})
			require.NoError(t, err)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var res models.UserResponse
			err = json.Unmarshal(bodyBytes, &res)
			require.NoError(t, err)

			require.Equal(t, c.expected.Username, res.Username)
			require.Equal(t, *c.expected.Bio, *res.Bio)
			require.Equal(t, *c.expected.ProfilePicture, *res.ProfilePicture)
			require.Equal(t, c.expected.CreatedAt.Unix(), res.CreatedAt.Unix())
			require.Equal(t, c.expected.VisitedFields, res.VisitedFields)
			require.Equal(t, c.expected.Winrate, res.Winrate)
			require.Equal(t, c.expected.FavoriteCity, res.FavoriteCity)
			require.Equal(t, c.expected.FavoriteSport, res.FavoriteSport)
			require.Equal(t, c.expected.FavoriteField, res.FavoriteField)
			require.Equal(t, c.expected.Sports, res.Sports)
			require.Equal(t, c.expected.Fields, res.Fields)
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
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			require.Equal(t, c.expectedCode, resp.StatusCode)

			if c.expected != nil {
				updated, err := s.db.GetUserById(r.Context(), c.urlUserId)
				require.NoError(t, err)

				require.Equal(t, c.expected.Username, updated.Username)
				require.Equal(t, c.expected.Email, updated.Email)
				require.Equal(t, *c.expected.Bio, *updated.Bio)
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
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

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
