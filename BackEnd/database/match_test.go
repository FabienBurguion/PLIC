package database

import (
	"PLIC/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDatabase_CreateMatch(t *testing.T) {
	type testCase struct {
		name  string
		match models.DBMatches
	}

	id := uuid.NewString()

	testCases := []testCase{
		{
			name: "Basic match creation",
			match: models.DBMatches{
				Id:              id,
				Sport:           models.Basket,
				Place:           "Paris",
				Date:            time.Now(),
				ParticipantNber: 8,
				CurrentState:    models.ManqueJoueur,
				Score1:          0,
				Score2:          0,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			s := &Service{}
			s.InitServiceTest()

			err := s.db.CreateMatch(ctx, c.match)
			require.NoError(t, err)

			dbMatch, err := s.db.GetMatchById(ctx, c.match.Id)
			require.NoError(t, err)
			require.NotNil(t, dbMatch)
			require.Equal(t, c.match.Id, dbMatch.Id)
			require.Equal(t, c.match.Place, dbMatch.Place)
		})
	}
}

func TestDatabase_GetMatchById(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
		matchID  string
		nilMatch bool
	}

	id := uuid.NewString()

	testCases := []testCase{
		{
			name: "Match exists",
			fixtures: DBFixtures{
				Matches: []models.DBMatches{
					{
						Id:              id,
						Sport:           models.Foot,
						Place:           "Lyon",
						Date:            time.Now(),
						ParticipantNber: 10,
						CurrentState:    models.ManqueJoueur,
						Score1:          1,
						Score2:          2,
					},
				},
			},
			matchID:  id,
			nilMatch: false,
		},
		{
			name:     "Match does not exist",
			matchID:  uuid.NewString(),
			nilMatch: true,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			match, err := s.db.GetMatchById(ctx, c.matchID)
			require.NoError(t, err)

			if c.nilMatch {
				require.Nil(t, match)
			} else {
				require.NotNil(t, match)
				require.Equal(t, c.matchID, match.Id)
			}
		})
	}
}

func TestDatabase_GetAllMatches(t *testing.T) {
	type testCase struct {
		name        string
		fixtures    DBFixtures
		expectedIDs []string
	}

	id1 := uuid.NewString()
	id2 := uuid.NewString()

	testCases := []testCase{
		{
			name: "Two matches exist",
			fixtures: DBFixtures{
				Matches: []models.DBMatches{
					{
						Id:              id1,
						Sport:           models.Foot,
						Place:           "Nice",
						Date:            time.Now(),
						ParticipantNber: 2,
						CurrentState:    models.ManqueJoueur,
						Score1:          0,
						Score2:          0,
					},
					{
						Id:              id2,
						Sport:           models.Basket,
						Place:           "Paris",
						Date:            time.Now(),
						ParticipantNber: 10,
						CurrentState:    models.ManqueJoueur,
						Score1:          0,
						Score2:          0,
					},
				},
			},
			expectedIDs: []string{id1, id2},
		},
		{
			name:        "No matches exist",
			fixtures:    DBFixtures{},
			expectedIDs: []string{},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			matches, err := s.db.GetAllMatches(ctx)
			require.NoError(t, err)

			require.GreaterOrEqual(t, len(matches), len(c.expectedIDs))

			for _, expectedID := range c.expectedIDs {
				var found bool
				for _, m := range matches {
					if m.Id == expectedID {
						found = true
						break
					}
				}
				require.True(t, found, "Expected match ID %s not found", expectedID)
			}
		})
	}
}
func TestDatabase_GetMatchesByUserID(t *testing.T) {

	type testCase struct {
		name             string
		fixtures         DBFixtures
		userID           string
		expectedMatchIDs []string
		expectError      bool
	}

	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	userID := uuid.NewString()
	bio := "Fan de foot et de basket"

	testCases := []testCase{
		{
			name: "User has two matches",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					{
						Id:        userID,
						Username:  "john_doe",
						Email:     "john@example.com",
						Bio:       &bio,
						Password:  "hashed-password",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
				Matches: []models.DBMatches{
					{
						Id:              matchID1,
						Sport:           models.Foot,
						Place:           "Lyon",
						Date:            time.Now(),
						ParticipantNber: 10,
						CurrentState:    models.ManqueJoueur,
						Score1:          1,
						Score2:          2,
					},
					{
						Id:              matchID2,
						Sport:           models.Basket,
						Place:           "Paris",
						Date:            time.Now(),
						ParticipantNber: 5,
						CurrentState:    models.Valide,
						Score1:          3,
						Score2:          3,
					},
				},
				UserMatches: []models.DBUserMatch{
					{UserID: userID, MatchID: matchID1},
					{UserID: userID, MatchID: matchID2},
				},
			},
			userID:           userID,
			expectedMatchIDs: []string{matchID1, matchID2},
			expectError:      false,
		},
		{
			name:             "User has no matches",
			fixtures:         DBFixtures{},
			userID:           uuid.NewString(),
			expectedMatchIDs: []string{},
			expectError:      false,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			matches, err := s.db.GetMatchesByUserID(ctx, c.userID)
			if c.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, len(c.expectedMatchIDs), len(matches))
			for i, m := range matches {
				require.Equal(t, c.expectedMatchIDs[i], m.Id)
			}
		})
	}
}

func TestDatabase_GetMatchCountByUserID(t *testing.T) {
	type testCase struct {
		name        string
		fixtures    DBFixtures
		userID      string
		expectedCnt int
		expectError bool
	}

	userID1 := uuid.NewString()
	userID2 := uuid.NewString()
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	matchID3 := uuid.NewString()
	matchID4 := uuid.NewString()
	bio := "Sportif"

	testCases := []testCase{
		{
			name: "User has two finished matches (Termine, Manque Score)",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					{
						Id:        userID1,
						Username:  "jean",
						Email:     "jean@example.com",
						Bio:       &bio,
						Password:  "password",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
				Matches: []models.DBMatches{
					{
						Id:              matchID1,
						Sport:           models.Foot,
						Place:           "Lyon",
						Date:            time.Now(),
						ParticipantNber: 10,
						CurrentState:    models.Termine,
						Score1:          2,
						Score2:          1,
					},
					{
						Id:              matchID2,
						Sport:           models.Basket,
						Place:           "Paris",
						Date:            time.Now(),
						ParticipantNber: 5,
						CurrentState:    models.ManqueScore,
						Score1:          3,
						Score2:          2,
					},
					{
						Id:              matchID3,
						Sport:           models.Foot,
						Place:           "Toulouse",
						Date:            time.Now(),
						ParticipantNber: 8,
						CurrentState:    models.Valide,
						Score1:          0,
						Score2:          0,
					},
				},
				UserMatches: []models.DBUserMatch{
					{UserID: userID1, MatchID: matchID1},
					{UserID: userID1, MatchID: matchID2},
					{UserID: userID1, MatchID: matchID3},
				},
			},
			userID:      userID1,
			expectedCnt: 2,
			expectError: false,
		},
		{
			name: "User has only non-finished matches",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					{Id: userID2, Username: "marie", Email: "marie@example.com"},
				},
				Matches: []models.DBMatches{
					{
						Id:              matchID4,
						Sport:           models.Basket,
						Place:           "Nice",
						Date:            time.Now(),
						ParticipantNber: 6,
						CurrentState:    models.Valide,
					},
				},
				UserMatches: []models.DBUserMatch{
					{UserID: userID2, MatchID: matchID4},
				},
			},
			userID:      userID2,
			expectedCnt: 0,
			expectError: false,
		},
		{
			name:        "User does not exist",
			fixtures:    DBFixtures{},
			userID:      uuid.NewString(),
			expectedCnt: 0,
			expectError: false,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			count, err := s.db.GetMatchCountByUserID(ctx, c.userID)
			if c.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.expectedCnt, count)
		})
	}
}
