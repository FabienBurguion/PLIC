package database

import (
	"PLIC/models"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateMatch(t *testing.T) {
	type testCase struct {
		name  string
		match models.DBMatches
	}

	s := &Service{}
	s.InitServiceTest()
	ctx := context.Background()

	court := models.NewDBCourtFixture().
		WithName("Test Court").
		WithLatitude(48.8566).
		WithLongitude(2.3522)

	err := s.db.InsertCourtForTest(ctx, court)
	require.NoError(t, err)

	matchID := uuid.NewString()
	testCases := []testCase{
		{
			name: "Basic match creation",
			match: models.DBMatches{
				Id:              matchID,
				Sport:           models.Basket,
				Date:            time.Now(),
				ParticipantNber: 8,
				CurrentState:    models.ManqueJoueur,
				Score1:          nil,
				Score2:          nil,
				CourtID:         court.Id,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := s.db.CreateMatch(ctx, c.match)
			require.NoError(t, err)

			dbMatch, err := s.db.GetMatchById(ctx, c.match.Id)
			require.NoError(t, err)
			require.NotNil(t, dbMatch)
			require.Equal(t, c.match.Id, dbMatch.Id)
			require.Equal(t, c.match.CourtID, dbMatch.CourtID)
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

	ctx := context.Background()
	s := &Service{}
	s.InitServiceTest()

	court := models.NewDBCourtFixture()

	id := uuid.NewString()

	testCases := []testCase{
		{
			name: "Match exists",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Matches: []models.DBMatches{
					{
						Id:              id,
						Sport:           models.Foot,
						Date:            time.Now(),
						ParticipantNber: 10,
						CurrentState:    models.ManqueJoueur,
						Score1:          ptr(1),
						Score2:          ptr(2),
						CourtID:         court.Id,
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

	// Créer un court pour les matchs
	court := models.NewDBCourtFixture()

	id1 := uuid.NewString()
	id2 := uuid.NewString()

	testCases := []testCase{
		{
			name: "Two matches exist",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Matches: []models.DBMatches{
					{
						Id:              id1,
						Sport:           models.Foot,
						Date:            time.Now(),
						ParticipantNber: 2,
						CurrentState:    models.ManqueJoueur,
						Score1:          nil,
						Score2:          nil,
						CourtID:         court.Id,
					},
					{
						Id:              id2,
						Sport:           models.Basket,
						Date:            time.Now(),
						ParticipantNber: 10,
						CurrentState:    models.ManqueJoueur,
						Score1:          nil,
						Score2:          nil,
						CourtID:         court.Id,
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
				Courts: courts, // Ajout des courts dans les fixtures
				Matches: []models.DBMatches{
					{
						Id:              matchID1,
						Sport:           models.Foot,
						Date:            time.Now(),
						ParticipantNber: 10,
						CurrentState:    models.ManqueJoueur,
						Score1:          ptr(1),
						Score2:          ptr(2),
						CourtID:         courtID,
					},
					{
						Id:              matchID2,
						Sport:           models.Basket,
						Date:            time.Now(),
						ParticipantNber: 5,
						CurrentState:    models.Valide,
						Score1:          ptr(3),
						Score2:          ptr(3),
						CourtID:         courtID,
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

			actualIDs := make([]string, len(matches))
			for i, m := range matches {
				actualIDs[i] = m.Id
			}

			require.ElementsMatch(t, c.expectedMatchIDs, actualIDs)
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
				Courts: courts,
				Matches: []models.DBMatches{
					{
						Id:              matchID1,
						Sport:           models.Foot,
						Date:            time.Now(),
						ParticipantNber: 10,
						CurrentState:    models.Termine,
						Score1:          ptr(2),
						Score2:          ptr(1),
						CourtID:         courtID,
					},
					{
						Id:              matchID2,
						Sport:           models.Basket,
						Date:            time.Now(),
						ParticipantNber: 5,
						CurrentState:    models.ManqueScore,
						Score1:          ptr(3),
						Score2:          ptr(2),
						CourtID:         courtID,
					},
					{
						Id:              matchID3,
						Sport:           models.Foot,
						Date:            time.Now(),
						ParticipantNber: 8,
						CurrentState:    models.Valide,
						Score1:          nil,
						Score2:          nil,
						CourtID:         courtID,
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
				Courts: courts,
				Matches: []models.DBMatches{
					{
						Id:              matchID4,
						Sport:           models.Basket,
						Date:            time.Now(),
						ParticipantNber: 6,
						CurrentState:    models.Valide,
						CourtID:         courtID,
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

func Test_GetMatchesByCourtId(t *testing.T) {
	court1 := models.NewDBCourtFixture()
	if court1.Id == "" {
		court1.Id = uuid.NewString()
	}
	court2 := models.NewDBCourtFixture()
	if court2.Id == "" {
		court2.Id = uuid.NewString()
	}

	match1 := models.DBMatches{
		Id:           uuid.NewString(),
		Sport:        models.Foot,
		Date:         time.Now().Add(-time.Hour),
		CurrentState: models.Termine,
		Score1:       ptr(3),
		Score2:       ptr(2),
		CourtID:      court1.Id,
	}
	match2 := models.DBMatches{
		Id:           uuid.NewString(),
		Sport:        models.Basket,
		Date:         time.Now(),
		CurrentState: models.Termine,
		Score1:       ptr(1),
		Score2:       ptr(1),
		CourtID:      court2.Id,
	}

	type testCase struct {
		name        string
		fixtures    DBFixtures
		courtID     string
		expectFound bool
		expectedLen int
		expectError bool
	}

	testCases := []testCase{
		{
			name: "Matches found for court1",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2},
			},
			courtID:     court1.Id,
			expectFound: true,
			expectedLen: 1,
			expectError: false,
		},
		{
			name: "Matches found for court2",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2},
			},
			courtID:     court2.Id,
			expectFound: true,
			expectedLen: 1,
			expectError: false,
		},
		{
			name: "No matches for unknown court",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2},
			},
			courtID:     uuid.NewString(),
			expectFound: false,
			expectedLen: 0,
			expectError: false,
		},
		{
			name: "No courts and no matches",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{},
				Matches: []models.DBMatches{},
			},
			courtID:     court1.Id,
			expectFound: false,
			expectedLen: 0,
			expectError: false,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()
			res, err := s.db.GetMatchesByCourtId(ctx, c.courtID)

			if c.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if c.expectFound {
				require.NotEmpty(t, res)
				require.Len(t, res, c.expectedLen)
				for _, m := range res {
					require.NotEmpty(t, m.Id)
				}
			} else {
				require.Empty(t, res)
			}
		})
	}
}

func TestDatabase_UpsertMatch(t *testing.T) {
	court := models.NewDBCourtFixture().
		WithName("Court Test").
		WithAddress("123 Test St")

	matchID := uuid.NewString()

	initialMatch := models.DBMatches{
		Id:              matchID,
		Sport:           models.Foot,
		Date:            time.Now(),
		ParticipantNber: 10,
		CurrentState:    models.ManqueJoueur,
		Score1:          ptr(0),
		Score2:          ptr(0),
		CourtID:         court.Id,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	updatedMatch := initialMatch
	updatedMatch.Score1 = ptr(5)
	updatedMatch.Score2 = ptr(3)
	updatedMatch.CurrentState = models.Termine

	type testCase struct {
		name          string
		fixtures      DBFixtures
		inputMatch    models.DBMatches
		expectedScore [2]*int
		expectedState models.MatchState
	}

	testCases := []testCase{
		{
			name: "Insert new match",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
			},
			inputMatch:    initialMatch,
			expectedScore: [2]*int{ptr(0), ptr(0)},
			expectedState: models.ManqueJoueur,
		},
		{
			name: "Update existing match",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{initialMatch},
			},
			inputMatch:    updatedMatch,
			expectedScore: [2]*int{ptr(5), ptr(3)},
			expectedState: models.Termine,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()
			err := s.db.UpsertMatch(ctx, c.inputMatch)
			require.NoError(t, err)

			matchFromDB, err := s.db.GetMatchById(ctx, c.inputMatch.Id)
			require.NoError(t, err)
			require.NotNil(t, matchFromDB)

			if c.expectedScore[0] == nil {
				require.Nil(t, matchFromDB.Score1)
			} else {
				require.NotNil(t, matchFromDB.Score1)
				require.Equal(t, *c.expectedScore[0], *matchFromDB.Score1)
			}
			if c.expectedScore[1] == nil {
				require.Nil(t, matchFromDB.Score2)
			} else {
				require.NotNil(t, matchFromDB.Score2)
				require.Equal(t, *c.expectedScore[1], *matchFromDB.Score2)
			}
			require.Equal(t, c.expectedState, matchFromDB.CurrentState)
		})
	}
}

func TestDatabase_CountUsersByMatchAndTeam(t *testing.T) {
	type testCase struct {
		name        string
		fixtures    DBFixtures
		matchID     string
		team        int
		expectedCnt int
		expectError bool
	}

	matchID := uuid.NewString()
	court := models.NewDBCourtFixture()
	teamA := 1
	teamB := 2

	user1 := models.DBUsers{Id: uuid.NewString(), Username: "user1", Email: "user1@example.com"}
	user2 := models.DBUsers{Id: uuid.NewString(), Username: "user2", Email: "user2@example.com"}
	user3 := models.DBUsers{Id: uuid.NewString(), Username: "user3", Email: "user3@example.com"}

	testCases := []testCase{
		{
			name: "2 users in team 1, 1 in team 2",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Users:  []models.DBUsers{user1, user2, user3},
				Matches: []models.DBMatches{
					{
						Id:              matchID,
						Sport:           models.Foot,
						Date:            time.Now(),
						ParticipantNber: 6,
						CurrentState:    models.Valide,
						CourtID:         court.Id,
					},
				},
				UserMatches: []models.DBUserMatch{
					{UserID: user1.Id, MatchID: matchID, Team: teamA},
					{UserID: user2.Id, MatchID: matchID, Team: teamA},
					{UserID: user3.Id, MatchID: matchID, Team: teamB},
				},
			},
			matchID:     matchID,
			team:        teamA,
			expectedCnt: 2,
			expectError: false,
		},
		{
			name: "No users in specified team",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Users:  []models.DBUsers{user1},
				Matches: []models.DBMatches{
					{Id: matchID, Sport: models.Foot, CourtID: court.Id, CurrentState: models.Valide},
				},
				UserMatches: []models.DBUserMatch{
					{UserID: user1.Id, MatchID: matchID, Team: 1},
				},
			},
			matchID:     matchID,
			team:        3,
			expectedCnt: 0,
			expectError: false,
		},
		{
			name:        "No such match",
			fixtures:    DBFixtures{},
			matchID:     uuid.NewString(),
			team:        1,
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

			count, err := s.db.CountUsersByMatchAndTeam(ctx, c.matchID, c.team)

			if c.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.expectedCnt, count)
		})
	}
}
