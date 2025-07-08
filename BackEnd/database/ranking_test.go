package database

import (
	"PLIC/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDatabase_GetRankedFieldsByUserID(t *testing.T) {
	type testCase struct {
		name           string
		fixtures       DBFixtures
		userID         string
		expectedFields []models.Field
		expectError    bool
	}

	userID := uuid.NewString()
	courtID1 := uuid.NewString()
	courtID2 := uuid.NewString()
	courtID3 := uuid.NewString()

	testCases := []testCase{
		{
			name: "User has rankings on multiple courts",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					{Id: userID, Username: "user", Email: "user@example.com", Password: "pwd"},
				},
				Courts: []models.DBCourt{
					{Id: courtID1, Name: "Central Park", Address: "NY", Latitude: 0.0, Longitude: 0.0},
					{Id: courtID2, Name: "Stade de Lyon", Address: "Lyon", Latitude: 0.0, Longitude: 0.0},
					{Id: courtID3, Name: "Playground", Address: "Paris", Latitude: 0.0, Longitude: 0.0},
				},
				Rankings: []models.DBRanking{
					{UserID: userID, CourtID: courtID1, Elo: 1200},
					{UserID: userID, CourtID: courtID2, Elo: 1300},
					{UserID: userID, CourtID: courtID3, Elo: 1250},
				},
			},
			userID: userID,
			expectedFields: []models.Field{
				{Ranking: 1, Name: "Stade de Lyon", Score: 1300},
				{Ranking: 2, Name: "Playground", Score: 1250},
				{Ranking: 3, Name: "Central Park", Score: 1200},
			},
			expectError: false,
		},
		{
			name: "User has no ranked courts",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					{Id: userID, Username: "empty", Email: "empty@example.com", Password: "pwd"},
				},
			},
			userID:         userID,
			expectedFields: []models.Field{},
			expectError:    false,
		},
		{
			name:           "Unknown user",
			fixtures:       DBFixtures{},
			userID:         uuid.NewString(),
			expectedFields: []models.Field{},
			expectError:    false,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()
			fields, err := s.db.GetRankedFieldsByUserID(ctx, c.userID)

			if c.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(c.expectedFields), len(fields))
			for i := range fields {
				require.Equal(t, c.expectedFields[i].Ranking, fields[i].Ranking)
				require.Equal(t, c.expectedFields[i].Name, fields[i].Name)
				require.Equal(t, c.expectedFields[i].Score, fields[i].Score)
			}
		})
	}
}

func TestDatabase_InsertRanking(t *testing.T) {
	s := &Service{}
	s.InitServiceTest()

	ctx := context.Background()

	userID := uuid.NewString()
	courtID := uuid.NewString()

	fixtures := DBFixtures{
		Users: []models.DBUsers{
			{Id: userID, Username: "bob", Email: "bob@example.com", Password: "pwd"},
		},
		Courts: []models.DBCourt{
			{
				Id:        courtID,
				Name:      "Central Park",
				Address:   "NY",
				Latitude:  0.0,
				Longitude: 0.0,
			},
		},
	}

	s.loadFixtures(fixtures)

	ranking := models.DBRanking{
		UserID:    userID,
		CourtID:   courtID,
		Elo:       1450,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("Insert new ranking", func(t *testing.T) {
		err := s.db.InsertRanking(ctx, ranking)
		require.NoError(t, err)

		var stored models.DBRanking
		err = s.db.Database.GetContext(ctx, &stored, `
			SELECT user_id, court_id, elo, created_at, updated_at
			FROM ranking
			WHERE user_id = $1 AND court_id = $2`,
			userID, courtID,
		)
		require.NoError(t, err)
		require.Equal(t, ranking.Elo, stored.Elo)
	})

	t.Run("Update existing ranking on conflict", func(t *testing.T) {
		ranking.Elo = 1600
		ranking.UpdatedAt = time.Now()

		err := s.db.InsertRanking(ctx, ranking)
		require.NoError(t, err)

		var updated models.DBRanking
		err = s.db.Database.GetContext(ctx, &updated, `
			SELECT user_id, court_id, elo
			FROM ranking
			WHERE user_id = $1 AND court_id = $2`,
			userID, courtID,
		)
		require.NoError(t, err)
		require.Equal(t, 1600, updated.Elo)
	})
}
