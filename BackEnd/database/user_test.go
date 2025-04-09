package database

import (
	"PLIC/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
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
			param: username,
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
			require.Equal(t, user.Username, c.param)
		})
	}
}
