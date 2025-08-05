package database

import (
	"PLIC/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
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

			err := s.db.CreateUser(ctx, models.DBUsers{
				Id:        id,
				Username:  "A name",
				Password:  "A password",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
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
	s := &Service{}
	s.InitServiceTest()

	userID := uuid.NewString()
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	courtID := uuid.NewString()

	courts := []models.DBCourt{
		{
			Id:        courtID,
			Name:      "Court central",
			Address:   "1 rue des sports",
			Longitude: 4.8357,
			Latitude:  45.7640,
			CreatedAt: time.Now(),
		},
	}

	fixtures := DBFixtures{
		Users: []models.DBUsers{
			{Id: userID, Username: "user", Email: "user@example.com", Password: "pwd"},
		},
		Courts: courts,
		Matches: []models.DBMatches{
			{Id: matchID1, Sport: models.Foot, Date: time.Now(), CurrentState: models.Termine, CourtID: courtID}, // Ajout CourtID
			{Id: matchID2, Sport: models.Foot, Date: time.Now(), CurrentState: models.Termine, CourtID: courtID}, // Ajout CourtID
		},
		UserMatches: []models.DBUserMatch{
			{UserID: userID, MatchID: matchID1},
			{UserID: userID, MatchID: matchID2},
		},
	}
	s.loadFixtures(fixtures)

	ctx := context.Background()
	field, err := s.db.GetFavoriteFieldByUserID(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, field)
	require.Equal(t, "Court central", *field)
}

func TestDatabase_GetFavoriteSportByUserID(t *testing.T) {
	s := &Service{}
	s.InitServiceTest()

	userID := uuid.NewString()
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	courtID := uuid.NewString()

	courts := []models.DBCourt{
		{
			Id:        courtID,
			Name:      "Court central",
			Address:   "10 avenue des sports",
			Longitude: 7.1234,
			Latitude:  43.5678,
			CreatedAt: time.Now(),
		},
	}

	fixtures := DBFixtures{
		Users: []models.DBUsers{
			{Id: userID, Username: "sporty", Email: "sporty@example.com", Password: "pwd"},
		},
		Courts: courts,
		Matches: []models.DBMatches{
			{Id: matchID1, Sport: models.Basket, Date: time.Now(), CurrentState: models.Termine, CourtID: courtID},
			{Id: matchID2, Sport: models.Basket, Date: time.Now(), CurrentState: models.Termine, CourtID: courtID},
		},
		UserMatches: []models.DBUserMatch{
			{UserID: userID, MatchID: matchID1},
			{UserID: userID, MatchID: matchID2},
		},
	}
	s.loadFixtures(fixtures)

	ctx := context.Background()
	sport, err := s.db.GetFavoriteSportByUserID(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, sport)
	require.Equal(t, models.Basket, *sport)
}

func TestDatabase_GetPlayedSportsByUserID(t *testing.T) {
	s := &Service{}
	s.InitServiceTest()

	userID := uuid.NewString()
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	matchID3 := uuid.NewString()
	courtID := uuid.NewString()

	courts := []models.DBCourt{
		{
			Id:        courtID,
			Name:      "Court central",
			Address:   "10 avenue des sports",
			Longitude: 4.8357,
			Latitude:  45.7640,
			CreatedAt: time.Now(),
		},
	}

	fixtures := DBFixtures{
		Users: []models.DBUsers{
			{Id: userID, Username: "multi", Email: "multi@example.com", Password: "pwd"},
		},
		Courts: courts,
		Matches: []models.DBMatches{
			{
				Id:           matchID1,
				Sport:        models.Foot,
				Date:         time.Now(),
				CurrentState: models.Termine,
				CourtID:      courtID,
			},
			{
				Id:           matchID2,
				Sport:        models.Basket,
				Date:         time.Now(),
				CurrentState: models.Termine,
				CourtID:      courtID,
			},
			{
				Id:           matchID3,
				Sport:        models.PingPong,
				Date:         time.Now(),
				CurrentState: models.Termine,
				CourtID:      courtID,
			},
		},
		UserMatches: []models.DBUserMatch{
			{UserID: userID, MatchID: matchID1},
			{UserID: userID, MatchID: matchID2},
			{UserID: userID, MatchID: matchID3},
		},
	}
	s.loadFixtures(fixtures)

	ctx := context.Background()
	sports, err := s.db.GetPlayedSportsByUserID(ctx, userID)

	require.NoError(t, err)
	require.ElementsMatch(t, []models.Sport{models.Foot, models.Basket, models.PingPong}, sports)
}
