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

	username := "Fabien"
	oldPassword := "password"
	newPassword := "<PASSWORD>"

	testCases := []testCase{
		{
			name: "Basic test",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithUsername(username).
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

			err = s.db.ChangePassword(ctx, username, string(hash))
			require.NoError(t, err)
			user, err := s.db.GetUserByUsername(ctx, username)
			require.NoError(t, err)
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword))
			require.NoError(t, err)
		})
	}
}

func TestDatabase_UpdateUser(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
		param    models.UserPatchRequest
		nilUser  bool
	}

	userId := uuid.NewString()

	username := "Fabien"
	email := "<EMAIL>"
	bio := "A bio"

	newUsername := "New username"
	newEmail := "<EMAIL2>"
	newBio := "A new bio"

	testCases := []testCase{
		{
			name: "Basic test",
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
				Bio:      ptr(newBio),
			},
			nilUser: false,
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
			nilUser: false,
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
			nilUser: false,
		},
		{
			name: "No users",
			param: models.UserPatchRequest{
				Username: ptr(newUsername),
				Email:    ptr(newEmail),
				Bio:      ptr(newBio),
			},
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

			err := s.db.UpdateUser(ctx, c.param, userId)
			require.NoError(t, err)
			user, err := s.db.GetUserById(ctx, userId)
			require.NoError(t, err)

			if c.nilUser {
				require.Nil(t, user)
				return
			}

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
		})
	}
}
