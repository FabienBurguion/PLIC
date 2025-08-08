package database

import (
	"PLIC/models"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateMatch(t *testing.T) {
	type testCase struct {
		name  string
		param models.DBMatches
	}

	court := models.NewDBCourtFixture().
		WithName("Test Court").
		WithLatitude(48.8566).
		WithLongitude(2.3522)

	matchID := uuid.NewString()
	testCases := []testCase{
		{
			name: "Basic param creation",
			param: models.NewDBMatchesFixture().
				WithCourtId(court.Id).
				WithId(matchID),
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := &Service{}
			s.InitServiceTest()
			ctx := context.Background()
			err := s.db.InsertCourtForTest(ctx, court)
			require.NoError(t, err)

			err = s.db.CreateMatch(ctx, c.param)
			require.NoError(t, err)

			dbMatch, err := s.db.GetMatchById(ctx, c.param.Id)
			require.NoError(t, err)
			require.NotNil(t, dbMatch)
			require.Equal(t, c.param.Id, dbMatch.Id)
			require.Equal(t, c.param.CourtID, dbMatch.CourtID)
		})
	}
}

func TestDatabase_GetMatchById(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected bool
	}

	court := models.NewDBCourtFixture()

	id := uuid.NewString()

	testCases := []testCase{
		{
			name: "Match exists",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(id).
						WithCourtId(court.Id),
				},
			},
			param:    id,
			expected: false,
		},
		{
			name:     "Match does not exist",
			param:    uuid.NewString(),
			expected: true,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			match, err := s.db.GetMatchById(ctx, c.param)
			require.NoError(t, err)

			if c.expected {
				require.Nil(t, match)
			} else {
				require.NotNil(t, match)
				require.Equal(t, c.param, match.Id)
			}
		})
	}
}

func TestDatabase_GetAllMatches(t *testing.T) {
	type testCase struct {
		name     string
		fixtures DBFixtures
		expected []string
	}

	court := models.NewDBCourtFixture()

	id1 := uuid.NewString()
	id2 := uuid.NewString()

	testCases := []testCase{
		{
			name: "Two matches exist",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(id1).
						WithCourtId(court.Id),
					models.NewDBMatchesFixture().
						WithId(id2).
						WithCourtId(court.Id),
				},
			},
			expected: []string{id1, id2},
		},
		{
			name:     "No matches exist",
			fixtures: DBFixtures{},
			expected: []string{},
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

			require.Equal(t, len(matches), len(c.expected))

			for _, expectedID := range c.expected {
				var found bool
				for _, m := range matches {
					if m.Id == expectedID {
						found = true
						break
					}
				}
				require.True(t, found, "Expected param ID %s not found", expectedID)
			}
		})
	}
}

func TestDatabase_GetMatchesByUserID(t *testing.T) {
	type expected struct {
		matchIDs []string
		isError  bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected expected
	}

	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	userID := uuid.NewString()

	courtID := uuid.NewString()
	courts := []models.DBCourt{
		models.NewDBCourtFixture().
			WithId(courtID),
	}

	testCases := []testCase{
		{
			name: "User has two matches",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userID),
				},
				Courts: courts,
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(matchID1).
						WithCourtId(courtID),
					models.NewDBMatchesFixture().
						WithId(matchID2).
						WithCourtId(courtID),
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
				matchIDs: []string{matchID1, matchID2},
				isError:  false,
			},
		},
		{
			name:  "User has no matches",
			param: uuid.NewString(),
			expected: expected{
				matchIDs: []string{},
				isError:  false,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			matches, err := s.db.GetMatchesByUserID(ctx, c.param)
			if c.expected.isError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, len(c.expected.matchIDs), len(matches))

			actualIDs := make([]string, len(matches))
			for i, m := range matches {
				actualIDs[i] = m.Id
			}

			require.ElementsMatch(t, c.expected.matchIDs, actualIDs)
		})
	}
}

func TestDatabase_GetMatchCountByUserID(t *testing.T) {
	type expected struct {
		cnt     int
		isError bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected expected
	}

	userID1 := uuid.NewString()
	userID2 := uuid.NewString()
	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	matchID3 := uuid.NewString()
	matchID4 := uuid.NewString()

	courtID := uuid.NewString()
	courts := []models.DBCourt{
		models.NewDBCourtFixture().
			WithId(courtID),
	}

	testCases := []testCase{
		{
			name: "User has two finished matches (Termine, Manque Score)",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userID1),
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
						WithCurrentState(models.ManqueScore),
					models.NewDBMatchesFixture().
						WithId(matchID3).
						WithCourtId(courtID).
						WithCurrentState(models.Valide),
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
			param: userID1,
			expected: expected{
				cnt:     2,
				isError: false,
			},
		},
		{
			name: "User has only non-finished matches",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userID2),
				},
				Courts: courts,
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(matchID4).
						WithCourtId(courtID).
						WithCurrentState(models.Valide),
				},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(userID2).
						WithMatchId(matchID4),
				},
			},
			param: userID2,
			expected: expected{
				cnt:     0,
				isError: false,
			},
		},
		{
			name:  "User does not exist",
			param: uuid.NewString(),
			expected: expected{
				cnt:     0,
				isError: false,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			count, err := s.db.GetMatchCountByUserID(ctx, c.param)
			if c.expected.isError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.expected.cnt, count)
		})
	}
}

func Test_GetMatchesByCourtId(t *testing.T) {
	type expected struct {
		found   bool
		len     int
		isError bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected expected
	}

	court1 := models.NewDBCourtFixture()
	court2 := models.NewDBCourtFixture()

	match1 := models.NewDBMatchesFixture().
		WithCourtId(court1.Id)
	match2 := models.NewDBMatchesFixture().
		WithCourtId(court2.Id)

	testCases := []testCase{
		{
			name: "Matches found for court1",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2},
			},
			param: court1.Id,
			expected: expected{
				found:   true,
				len:     1,
				isError: false,
			},
		},
		{
			name: "Matches found for court2",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2},
			},
			param: court2.Id,
			expected: expected{
				found:   true,
				len:     1,
				isError: false,
			},
		},
		{
			name: "No matches for unknown court",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court1, court2},
				Matches: []models.DBMatches{match1, match2},
			},
			param: uuid.NewString(),
			expected: expected{
				found:   false,
				len:     0,
				isError: false,
			},
		},
		{
			name: "No courts and no matches",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{},
				Matches: []models.DBMatches{},
			},
			param: court1.Id,
			expected: expected{
				found:   false,
				len:     0,
				isError: false,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()
			res, err := s.db.GetMatchesByCourtId(ctx, c.param)

			if c.expected.isError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if c.expected.found {
				require.NotEmpty(t, res)
				require.Len(t, res, c.expected.len)
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
	type expected struct {
		score [2]*int
		state models.MatchState
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    models.DBMatches
		expected expected
	}

	court := models.NewDBCourtFixture().
		WithName("Court Test").
		WithAddress("123 Test St")

	matchID := uuid.NewString()

	initialMatch := models.NewDBMatchesFixture().
		WithId(matchID).
		WithCourtId(court.Id).
		WithScore1(0).
		WithScore2(0)

	updatedMatch := initialMatch
	updatedMatch.Score1 = ptr(5)
	updatedMatch.Score2 = ptr(3)
	updatedMatch.CurrentState = models.Termine

	testCases := []testCase{
		{
			name: "Insert new param",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
			},
			param: initialMatch,
			expected: expected{
				score: [2]*int{ptr(0), ptr(0)},
				state: models.ManqueJoueur,
			},
		},
		{
			name: "Update existing param",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{initialMatch},
			},
			param: updatedMatch,
			expected: expected{
				score: [2]*int{ptr(5), ptr(3)},
				state: models.Termine,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()
			err := s.db.UpsertMatch(ctx, c.param)
			require.NoError(t, err)

			matchFromDB, err := s.db.GetMatchById(ctx, c.param.Id)
			require.NoError(t, err)
			require.NotNil(t, matchFromDB)

			if c.expected.score[0] == nil {
				require.Nil(t, matchFromDB.Score1)
			} else {
				require.NotNil(t, matchFromDB.Score1)
				require.Equal(t, *c.expected.score[0], *matchFromDB.Score1)
			}
			if c.expected.score[1] == nil {
				require.Nil(t, matchFromDB.Score2)
			} else {
				require.NotNil(t, matchFromDB.Score2)
				require.Equal(t, *c.expected.score[1], *matchFromDB.Score2)
			}
			require.Equal(t, c.expected.state, matchFromDB.CurrentState)
		})
	}
}

func TestDatabase_CountUsersByMatchAndTeam(t *testing.T) {
	type param struct {
		matchID string
		team    int
	}

	type expected struct {
		cnt     int
		IsError bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    param
		expected expected
	}

	matchID := uuid.NewString()
	court := models.NewDBCourtFixture()
	teamA := 1
	teamB := 2

	user1 := models.NewDBUsersFixture().
		WithUsername("user1").
		WithEmail("email1")
	user2 := models.NewDBUsersFixture().
		WithUsername("user2").
		WithEmail("email2")
	user3 := models.NewDBUsersFixture().
		WithUsername("user3").
		WithEmail("email3")

	testCases := []testCase{
		{
			name: "2 users in team 1, 1 in team 2",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Users:  []models.DBUsers{user1, user2, user3},
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(matchID).
						WithCourtId(court.Id),
				},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user1.Id).
						WithMatchId(matchID).
						WithTeam(teamA),
					models.NewDBUserMatchFixture().
						WithUserId(user2.Id).
						WithMatchId(matchID).
						WithTeam(teamA),
					models.NewDBUserMatchFixture().
						WithUserId(user3.Id).
						WithMatchId(matchID).
						WithTeam(teamB),
				},
			},
			param: param{
				matchID: matchID,
				team:    teamA,
			},
			expected: expected{
				cnt:     2,
				IsError: false,
			},
		},
		{
			name: "No users in specified team",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
				Users:  []models.DBUsers{user1},
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(matchID).
						WithCourtId(court.Id),
				},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user1.Id).
						WithMatchId(matchID).
						WithTeam(1),
				},
			},
			param: param{
				matchID: matchID,
				team:    3,
			},
			expected: expected{
				cnt:     0,
				IsError: false,
			},
		},
		{
			name:     "No such param",
			fixtures: DBFixtures{},
			param: param{
				matchID: uuid.NewString(),
				team:    1,
			},
			expected: expected{
				cnt:     0,
				IsError: false,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			s := &Service{}
			s.InitServiceTest()
			s.loadFixtures(c.fixtures)

			count, err := s.db.CountUsersByMatchAndTeam(ctx, c.param.matchID, c.param.team)

			if c.expected.IsError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.expected.cnt, count)
		})
	}
}
