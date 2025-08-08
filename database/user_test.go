package database

import (
	"PLIC/models"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestDatabase_CheckUserExist(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
		expected bool
	}
	ctx := context.Background()

	userId := uuid.NewString()

	testCases := []testCase{
		{
			name: "User exists",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId),
				},
			},
			expected: true,
		},
		{
			name: "User does not exist",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture(),
				},
			},
			expected: false,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			res, err := s.db.CheckUserExist(ctx, userId)
			require.NoError(t, err)
			require.Equal(t, c.expected, res)
		})
	}
}

func TestDatabase_CreateUser(t *testing.T) {
	type testCase struct {
		name  string
		param string
	}

	id := uuid.NewString()

	testCases := []testCase{
		{
			name:  "Basic test",
			param: id,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()

			ctx := context.Background()

			err := s.db.CreateUser(ctx, models.NewDBUsersFixture().
				WithId(id).
				WithUsername("A name"),
			)
			require.NoError(t, err)
			user, err := s.db.GetUserByUsername(ctx, "A name")
			require.NoError(t, err)
			require.Equal(t, user.Id, c.param)
		})
	}
}

func TestDatabase_GetUserByUsername(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		nilUser  bool
	}

	username := "Fabien"

	testCases := []testCase{
		{
			name: "Basic test",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithUsername(username),
				},
			},
			param:   username,
			nilUser: false,
		},
		{
			name:    "User does not exist",
			param:   username,
			nilUser: true,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()

			user, err := s.db.GetUserByUsername(ctx, c.param)
			require.NoError(t, err)
			if c.nilUser {
				require.Nil(t, user)
			} else {
				require.Equal(t, user.Username, c.param)
			}
		})
	}
}

func TestDatabase_GetUserByEmail(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		nilUser  bool
	}

	email := "fabien@example.com"

	testCases := []testCase{
		{
			name: "Basic test",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithEmail(email),
				},
			},
			param: email,
		},
		{
			name:    "User does not exist",
			param:   email,
			nilUser: true,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()

			user, err := s.db.GetUserByEmail(ctx, c.param)
			require.NoError(t, err)
			if c.nilUser {
				require.Nil(t, user)
			} else {
				require.Equal(t, user.Email, c.param)
			}
		})
	}
}

func TestDatabase_ChangePassword(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
	}

	email := "email@email.com"
	oldPassword := "password"
	newPassword := "<PASSWORD>"

	testCases := []testCase{
		{
			name: "Basic test",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithEmail(email).
						WithPassword(oldPassword),
				},
			},
			param: newPassword,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()

			hash, err := bcrypt.GenerateFromPassword([]byte(c.param), bcrypt.DefaultCost)
			require.NoError(t, err)

			err = s.db.ChangePassword(ctx, email, string(hash))
			require.NoError(t, err)
			user, err := s.db.GetUserByEmail(ctx, email)
			require.NoError(t, err)
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword))
			require.NoError(t, err)
		})
	}
}

func TestDatabase_UpdateUser(t *testing.T) {
	type testCase struct {
		name          string
		fixtures      DBFixtures
		param         models.UserPatchRequest
		newUpdateTime time.Time
		nilUser       bool
	}

	userId := uuid.NewString()

	username := "Fabien"
	email := "<EMAIL>"
	bio := "A bio"

	newUsername := "New username"
	newEmail := "<EMAIL2>"
	newBio := "A new bio"

	newFieldId := uuid.NewString()

	newUpdatedTime := time.Now().UTC().Add(time.Hour)

	testCases := []testCase{
		{
			name: "Basic test",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithUsername(username).
						WithEmail(email).
						WithBio(bio).
						WitUpdatedAt(time.Now()),
				},
			},
			param: models.UserPatchRequest{
				Username: ptr(newUsername),
				Email:    ptr(newEmail),
				Bio:      ptr(newBio),
			},
			newUpdateTime: newUpdatedTime,
			nilUser:       false,
		},
		{
			name: "Bio only",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithUsername(username).
						WithEmail(email).
						WithBio(bio),
				},
			},
			param: models.UserPatchRequest{
				Username: nil,
				Email:    nil,
				Bio:      ptr(newBio),
			},
			newUpdateTime: newUpdatedTime,
			nilUser:       false,
		},
		{
			name: "Username + email",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithUsername(username).
						WithEmail(email).
						WithBio(bio),
				},
			},
			param: models.UserPatchRequest{
				Username: ptr(newUsername),
				Email:    ptr(newEmail),
				Bio:      nil,
			},
			newUpdateTime: newUpdatedTime,
			nilUser:       false,
		},
		{
			name: "No users",
			param: models.UserPatchRequest{
				Username: ptr(newUsername),
				Email:    ptr(newEmail),
				Bio:      ptr(newBio),
			},
			newUpdateTime: newUpdatedTime,
			nilUser:       true,
		},
		{
			name: "Current field id",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userId).
						WithUsername(username).
						WithEmail(email).
						WithBio(bio),
				},
			},
			param: models.UserPatchRequest{
				CurrentFieldId: ptr(newFieldId),
			},
			newUpdateTime: newUpdatedTime,
			nilUser:       false,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()

			err := s.db.UpdateUser(ctx, c.param, userId, c.newUpdateTime)
			require.NoError(t, err)
			user, err := s.db.GetUserById(ctx, userId)
			require.NoError(t, err)

			if c.nilUser {
				require.Nil(t, user)
				return
			}

			require.WithinDuration(t, c.newUpdateTime, user.UpdatedAt, time.Millisecond,
				"expected UpdatedAt to be within 1ms of %v, got %v", c.newUpdateTime, user.UpdatedAt)

			if c.param.Username != nil {
				require.Equal(t, user.Username, newUsername)
			} else {
				require.Equal(t, user.Username, username)
			}
			if c.param.Email != nil {
				require.Equal(t, user.Email, newEmail)
			} else {
				require.Equal(t, user.Email, email)
			}
			if c.param.Bio != nil {
				require.Equal(t, *user.Bio, newBio)
			} else {
				require.Equal(t, *user.Bio, bio)
			}
			if c.param.CurrentFieldId != nil {
				require.Equal(t, *user.CurrentFieldId, newFieldId)
			} else {
				require.Nil(t, user.CurrentFieldId)
			}
		})
	}
}

func TestDatabase_DeleteUser(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
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
		},
		{
			name: "User doesn't exist",
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()

			err := s.db.DeleteUser(ctx, userId)
			require.NoError(t, err)
			user, err := s.db.GetUserById(ctx, userId)
			require.NoError(t, err)
			require.Nil(t, user)
		})
	}
}

func TestDatabase_GetFavoriteFieldByUserID(t *testing.T) {
	type expected struct {
		IsError bool
		res     *string
	}

	type testCase struct {
		name        string
		fixtures    DBFixtures
		inputUserID string
		expected    expected
	}

	courtID := uuid.NewString()
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	userID := uuid.NewString()

	courts := []models.DBCourt{
		models.NewDBCourtFixture().
			WithId(courtID).
			WithName("Court central"),
	}

	testCases := []testCase{
		{
			name: "User with favorite court",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userID),
				},
				Courts: courts,
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(matchID1).
						WithCourtId(courtID).
						WithCurrentState(models.Termine),
					models.NewDBMatchesFixture().
						WithId(matchID2).
						WithCourtId(courtID).
						WithCurrentState(models.Termine),
				},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID1),
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID2),
				},
			},
			inputUserID: userID,
			expected: expected{
				IsError: false,
				res:     ptr("Court central"),
			},
		},
		{
			name: "User with no matches",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture(),
				},
			},
			inputUserID: "empty-user",
			expected: expected{
				IsError: false,
				res:     nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(tc.fixtures)

			ctx := context.Background()
			result, err := s.db.GetFavoriteFieldByUserID(ctx, tc.inputUserID)

			if tc.expected.IsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.expected.res == nil {
					require.Nil(t, result)
				} else {
					require.NotNil(t, result)
					require.Equal(t, *tc.expected.res, *result)
				}
			}
		})
	}
}

func TestDatabase_GetFavoriteSportByUserID(t *testing.T) {
	type expected struct {
		IsError bool
		res     *models.Sport
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected expected
	}

	userID := uuid.NewString()
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	courtID := uuid.NewString()

	courts := []models.DBCourt{
		models.NewDBCourtFixture().
			WithId(courtID).
			WithName("Court central"),
	}

	testCases := []testCase{
		{
			name: "User with favorite sport",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userID),
				},
				Courts: courts,
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(matchID1).
						WithCourtId(courtID).
						WithSport(models.Basket).
						WithCurrentState(models.Termine),
					models.NewDBMatchesFixture().
						WithId(matchID2).
						WithCourtId(courtID).
						WithSport(models.Basket).
						WithCurrentState(models.Termine),
				},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID1),
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID2),
				},
			},
			param: userID,
			expected: expected{
				IsError: false,
				res:     ptr(models.Basket),
			},
		},
		{
			name: "User with no matches",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture(),
				},
			},
			param: "empty-user",
			expected: expected{
				IsError: false,
				res:     nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(tc.fixtures)

			ctx := context.Background()
			result, err := s.db.GetFavoriteSportByUserID(ctx, tc.param)

			if tc.expected.IsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.expected.res == nil {
					require.Nil(t, result)
				} else {
					require.NotNil(t, result)
					require.Equal(t, *tc.expected.res, *result)
				}
			}
		})
	}
}

func TestDatabase_GetPlayedSportsByUserID(t *testing.T) {
	type expected struct {
		IsError bool
		sports  []models.Sport
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected expected
	}

	userID := uuid.NewString()
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	matchID3 := uuid.NewString()
	courtID := uuid.NewString()

	testCases := []testCase{
		{
			name: "User with multiple sports",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userID),
				},
				Courts: []models.DBCourt{
					models.NewDBCourtFixture().
						WithId(courtID),
				},
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(matchID1).
						WithSport(models.Foot).
						WithCourtId(courtID).
						WithCurrentState(models.Termine),
					models.NewDBMatchesFixture().
						WithId(matchID2).
						WithSport(models.Basket).
						WithCourtId(courtID).
						WithCurrentState(models.Termine),
					models.NewDBMatchesFixture().
						WithId(matchID3).
						WithSport(models.PingPong).
						WithCourtId(courtID).
						WithCurrentState(models.Termine),
				},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID1),
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID2),
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID3),
				},
			},
			param: userID,
			expected: expected{
				IsError: false,
				sports:  []models.Sport{models.Foot, models.Basket, models.PingPong},
			},
		},
		{
			name: "User with no matches",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture(),
				},
			},
			param: "unknown-user",
			expected: expected{
				IsError: false,
				sports:  []models.Sport{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(tc.fixtures)

			ctx := context.Background()
			sports, err := s.db.GetPlayedSportsByUserID(ctx, tc.param)

			if tc.expected.IsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, tc.expected.sports, sports)
			}
		})
	}
}
