package database

import (
	"PLIC/models"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDatabase_InsertTerrain(t *testing.T) {
	type testCase struct {
		name  string
		param string
	}

	testLat := 48.8566
	testLng := 2.3522

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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			ctx := context.Background()

			err := s.db.InsertCourt(ctx, id, models.Place{
				Name:    "Test Court",
				Address: "123 Rue Test",
				Geometry: struct {
					Location struct {
						Lat float64 `json:"lat"`
						Lng float64 `json:"lng"`
					} `json:"location"`
				}{
					Location: struct {
						Lat float64 `json:"lat"`
						Lng float64 `json:"lng"`
					}{
						Lat: testLat,
						Lng: testLng,
					},
				},
			}, time.Now())
			require.NoError(t, err)
			court, err := s.db.GetTerrainByAddress(ctx, "123 Rue Test")
			require.NoError(t, err)
			require.Equal(t, court.Id, c.param)
		})
	}
}

func TestDatabase_GetAllCourts(t *testing.T) {
	type testCase struct {
		name     string
		expected models.DBCourt
	}

	testID := uuid.NewString()

	testCases := []testCase{
		{
			name: "Basic terrain retrieval",
			expected: models.DBCourt{
				Id:        testID,
				Address:   "123 Rue Test",
				Longitude: 2.3522,
				Latitude:  48.8566,
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
			ctx := context.Background()

			err := s.db.InsertCourt(ctx, c.expected.Id, models.Place{
				Name:    c.expected.Name,
				Address: c.expected.Address,
				Geometry: struct {
					Location struct {
						Lat float64 `json:"lat"`
						Lng float64 `json:"lng"`
					} `json:"location"`
				}{
					Location: struct {
						Lat float64 `json:"lat"`
						Lng float64 `json:"lng"`
					}{
						Lat: c.expected.Latitude,
						Lng: c.expected.Longitude,
					},
				},
			}, c.expected.CreatedAt)
			require.NoError(t, err)

			terrains, err := s.db.GetAllCourts(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, terrains)

			var found bool
			for _, terr := range terrains {
				if terr.Id == c.expected.Id {
					require.Equal(t, c.expected.Address, terr.Address)
					require.Equal(t, c.expected.Longitude, terr.Longitude)
					require.Equal(t, c.expected.Latitude, terr.Latitude)
					require.Equal(t, c.expected.Name, terr.Name)
					found = true
					break
				}
			}
			require.True(t, found, "terrain not found in results")
		})
	}
}

func TestDatabase_GetVisitedFieldCountByUserID(t *testing.T) {
	type testCase struct {
		name          string
		fixtures      DBFixtures
		userID        string
		expectedCount int
		expectError   bool
	}

	userID1 := uuid.NewString()
	court1 := models.NewDBCourtFixture().
		WithAddress("12 rue de Paris, Paris")
	court2 := models.NewDBCourtFixture().
		WithAddress("5 avenue des Champs, Lyon")
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	matchID3 := uuid.NewString()

	testCases := []testCase{
		{
			name: "User has matches in different places",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court1, court2},
				Users: []models.DBUsers{
					models.NewDBUsersFixture().WithId(userID1),
				},
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(matchID1).
						WithCourtId(court1.Id),
					models.NewDBMatchesFixture().
						WithId(matchID2).
						WithCourtId(court2.Id),
					models.NewDBMatchesFixture().
						WithId(matchID3).
						WithCourtId(court1.Id),
				},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(userID1).
						WithMatchId(matchID1),
					models.NewDBUserMatchFixture().
						WithUserId(userID1).
						WithMatchId(matchID2),
					models.NewDBUserMatchFixture().
						WithUserId(userID1).
						WithMatchId(matchID3),
				},
			},
			userID:        userID1,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "User has no matches",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userID1),
				},
			},
			userID:        userID1,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Unknown user",
			fixtures:      DBFixtures{},
			userID:        uuid.NewString(),
			expectedCount: 0,
			expectError:   false,
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

			ctx := context.Background()
			count, err := s.db.GetVisitedFieldCountByUserID(ctx, c.userID)

			if c.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, c.expectedCount, count)
		})
	}
}

func TestDatabase_GetCourtByID(t *testing.T) {
	type expected struct {
		found bool
		name  string
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected expected
	}

	court1 := models.NewDBCourtFixture().WithName("Court One")
	court2 := models.NewDBCourtFixture().WithName("Court Two")

	testCases := []testCase{
		{
			name: "Court exists",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court1, court2},
			},
			param: court1.Id,
			expected: expected{
				found: true,
				name:  "Court One",
			},
		},
		{
			name: "Court does not exist",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court1},
			},
			param: uuid.NewString(),
			expected: expected{
				found: false,
			},
		},
		{
			name:     "Empty database",
			fixtures: DBFixtures{},
			param:    uuid.NewString(),
			expected: expected{
				found: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()

			s.loadFixtures(tc.fixtures)

			got, err := s.db.GetCourtByID(ctx, tc.param)
			require.NoError(t, err)

			if !tc.expected.found {
				require.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			require.Equal(t, tc.param, got.Id)
			require.Equal(t, tc.expected.name, got.Name)
		})
	}
}
