package database

import (
	"PLIC/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDatabase_InsertTerrain(t *testing.T) {
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

	testLat := 48.8566
	testLng := 2.3522

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
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
			s.InitServiceTest()
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
	court1 := models.NewDBCourtFixture().WithAddress("12 rue de Paris, Paris")
	court2 := models.NewDBCourtFixture().WithAddress("5 avenue des Champs, Lyon")
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	matchID3 := uuid.NewString()

	testCases := []testCase{
		{
			name: "User has matches in different places",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court1, court2},
				Users: []models.DBUsers{
					{Id: userID1, Username: "toto", Email: "toto@example.com", Password: "xxx"},
				},
				Matches: []models.DBMatches{
					{
						Id:              matchID1,
						Sport:           models.Foot,
						Date:            time.Now(),
						ParticipantNber: 10,
						CurrentState:    models.Valide,
						Score1:          1,
						Score2:          2,
						CourtID:         court1.Id,
					},
					{
						Id:              matchID2,
						Sport:           models.Basket,
						Date:            time.Now(),
						ParticipantNber: 8,
						CurrentState:    models.Termine,
						Score1:          3,
						Score2:          3,
						CourtID:         court2.Id,
					},
					{
						Id:              matchID3,
						Sport:           models.Foot,
						Date:            time.Now(),
						ParticipantNber: 12,
						CurrentState:    models.Valide,
						Score1:          0,
						Score2:          0,
						CourtID:         court1.Id,
					},
				},
				UserMatches: []models.DBUserMatch{
					{UserID: userID1, MatchID: matchID1},
					{UserID: userID1, MatchID: matchID2},
					{UserID: userID1, MatchID: matchID3},
				},
			},
			userID:        userID1,
			expectedCount: 2, // Paris and Lyon (two distinct Place fields)
			expectError:   false,
		},
		{
			name: "User has no matches",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					{Id: userID1, Username: "tata", Email: "tata@example.com", Password: "xxx"},
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
			s.InitServiceTest()
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
