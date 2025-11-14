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
		name     string
		fixtures DBFixtures
		param    models.DBMatches
	}
	user := models.NewDBUsersFixture()

	court := models.NewDBCourtFixture().
		WithName("Test Court").
		WithLatitude(48.8566).
		WithLongitude(2.3522)

	matchID := uuid.NewString()
	testCases := []testCase{
		{
			name: "Basic param creation",
			fixtures: DBFixtures{
				Users: []models.DBUsers{user},
			},
			param: models.NewDBMatchesFixture().
				WithCourtId(court.Id).
				WithId(matchID).
				WithCreatorId(user.Id),
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
	user := models.NewDBUsersFixture()

	court := models.NewDBCourtFixture()

	id := uuid.NewString()

	testCases := []testCase{
		{
			name: "Match exists",
			fixtures: DBFixtures{
				Users:  []models.DBUsers{user},
				Courts: []models.DBCourt{court},
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(id).
						WithCourtId(court.Id).
						WithCreatorId(user.Id),
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
	user := models.NewDBUsersFixture()

	court := models.NewDBCourtFixture()

	id1 := uuid.NewString()
	id2 := uuid.NewString()

	testCases := []testCase{
		{
			name: "Two matches exist",
			fixtures: DBFixtures{
				Users:  []models.DBUsers{user},
				Courts: []models.DBCourt{court},
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(id1).
						WithCourtId(court.Id).
						WithCreatorId(user.Id),
					models.NewDBMatchesFixture().
						WithId(id2).
						WithCourtId(court.Id).
						WithCreatorId(user.Id),
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
						WithCourtId(courtID).
						WithCreatorId(userID),
					models.NewDBMatchesFixture().
						WithId(matchID2).
						WithCourtId(courtID).
						WithCreatorId(userID),
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
						WithCurrentState(models.Termine).
						WithCreatorId(userID1),
					models.NewDBMatchesFixture().
						WithId(matchID2).
						WithCourtId(courtID).
						WithCurrentState(models.ManqueScore).
						WithCreatorId(userID1),
					models.NewDBMatchesFixture().
						WithId(matchID3).
						WithCourtId(courtID).
						WithCurrentState(models.Valide).
						WithCreatorId(userID1),
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
						WithCurrentState(models.Valide).
						WithCreatorId(userID2),
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
	user := models.NewDBUsersFixture()

	court1 := models.NewDBCourtFixture()
	court2 := models.NewDBCourtFixture()

	match1 := models.NewDBMatchesFixture().
		WithCourtId(court1.Id).
		WithCreatorId(user.Id)
	match2 := models.NewDBMatchesFixture().
		WithCourtId(court2.Id).
		WithCreatorId(user.Id)

	testCases := []testCase{
		{
			name: "Matches found for court1",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{user},
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
				Users:   []models.DBUsers{user},
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
				Users:   []models.DBUsers{user},
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
	user := models.NewDBUsersFixture()

	court := models.NewDBCourtFixture().
		WithName("Court Test").
		WithAddress("123 Test St")

	matchID := uuid.NewString()

	initialMatch := models.NewDBMatchesFixture().
		WithId(matchID).
		WithCourtId(court.Id).
		WithScore1(0).
		WithScore2(0).
		WithCreatorId(user.Id)

	updatedMatch := initialMatch
	updatedMatch.Score1 = ptr(5)
	updatedMatch.Score2 = ptr(3)
	updatedMatch.CurrentState = models.Termine

	testCases := []testCase{
		{
			name: "Insert new param",
			fixtures: DBFixtures{
				Users:  []models.DBUsers{user},
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
				Users:   []models.DBUsers{user},
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()
			err := s.db.UpsertMatch(ctx, c.param, time.Now())
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
						WithCourtId(court.Id).
						WithCreatorId(user1.Id),
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
						WithCourtId(court.Id).
						WithCreatorId(user1.Id),
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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

func TestDatabase_GetUserInMatch(t *testing.T) {
	type expected struct {
		found bool
		team  int
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		userID   string
		matchID  string
		expected expected
	}

	user1 := models.NewDBUsersFixture().WithUsername("user1").WithEmail("email1")
	user2 := models.NewDBUsersFixture().WithUsername("user2").WithEmail("email2")
	court := models.NewDBCourtFixture()
	match1 := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCreatorId(user1.Id)
	match2 := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCreatorId(user1.Id)

	testCases := []testCase{
		{
			name: "User is in the match",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{user1, user2},
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match1, match2},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user1.Id).
						WithMatchId(match1.Id).
						WithTeam(1),
				},
			},
			userID:  user1.Id,
			matchID: match1.Id,
			expected: expected{
				found: true,
				team:  1,
			},
		},
		{
			name: "User not in the match (different user)",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{user1, user2},
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match1},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user1.Id).
						WithMatchId(match1.Id).
						WithTeam(2),
				},
			},
			userID:  user2.Id,
			matchID: match1.Id,
			expected: expected{
				found: false,
			},
		},
		{
			name: "Unknown match id",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{user1},
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match1},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(user1.Id).
						WithMatchId(match1.Id).
						WithTeam(1),
				},
			},
			userID:  user1.Id,
			matchID: uuid.NewString(),
			expected: expected{
				found: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
			s.loadFixtures(tc.fixtures)

			ctx := context.Background()
			got, err := s.db.GetUserInMatch(ctx, tc.userID, tc.matchID)
			require.NoError(t, err)

			if !tc.expected.found {
				require.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			require.Equal(t, tc.userID, got.UserID)
			require.Equal(t, tc.matchID, got.MatchID)
			require.Equal(t, tc.expected.team, got.Team)
		})
	}
}

func TestDatabase_GetUserMatchesByMatchID(t *testing.T) {
	type expected struct {
		wantLen int
		userIDs []string
		teams   []int
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		matchID  string
		expected expected
	}

	court := models.NewDBCourtFixture()

	u1 := models.NewDBUsersFixture().WithUsername("u1").WithEmail("u1@example.com")
	u2 := models.NewDBUsersFixture().WithUsername("u2").WithEmail("u2@example.com")
	u3 := models.NewDBUsersFixture().WithUsername("u3").WithEmail("u3@example.com")

	match1 := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCreatorId(u1.Id)
	match2 := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCreatorId(u1.Id)

	testCases := []testCase{
		{
			name: "User matches found for match1",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match1, match2},
				Users:   []models.DBUsers{u1, u2, u3},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(u1.Id).WithMatchId(match1.Id).WithTeam(1),
					models.NewDBUserMatchFixture().WithUserId(u2.Id).WithMatchId(match1.Id).WithTeam(1),
					models.NewDBUserMatchFixture().WithUserId(u3.Id).WithMatchId(match1.Id).WithTeam(2),
					models.NewDBUserMatchFixture().WithUserId(u1.Id).WithMatchId(match2.Id).WithTeam(2),
				},
			},
			matchID: match1.Id,
			expected: expected{
				wantLen: 3,
				userIDs: []string{u1.Id, u2.Id, u3.Id},
				teams:   []int{1, 1, 2},
			},
		},
		{
			name: "No user matches for match2 (only one unrelated row)",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match1, match2},
				Users:   []models.DBUsers{u1},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(u1.Id).WithMatchId(match1.Id).WithTeam(1),
				},
			},
			matchID: match2.Id,
			expected: expected{
				wantLen: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
			s.loadFixtures(tc.fixtures)

			ctx := context.Background()
			got, err := s.db.GetUserMatchesByMatchID(ctx, tc.matchID)
			require.NoError(t, err)

			require.Len(t, got, tc.expected.wantLen)

			if tc.expected.wantLen > 0 {
				gotUserIDs := make([]string, len(got))
				gotTeams := make([]int, len(got))
				for i, um := range got {
					gotUserIDs[i] = um.UserID
					gotTeams[i] = um.Team
					require.Equal(t, tc.matchID, um.MatchID)
				}
				require.ElementsMatch(t, tc.expected.userIDs, gotUserIDs)
				require.ElementsMatch(t, tc.expected.teams, gotTeams)
			}
		})
	}
}

func TestDatabase_GetUserWinrate(t *testing.T) {
	ptr := func(i int) *int { return &i }

	type expected struct {
		wantNil bool
		value   int
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		userID   string
		expected expected
	}

	court := models.NewDBCourtFixture()

	u1 := models.NewDBUsersFixture().WithUsername("u1").WithEmail("u1@example.com")
	u2 := models.NewDBUsersFixture().WithUsername("u2").WithEmail("u2@example.com")
	u3 := models.NewDBUsersFixture().WithUsername("u3").WithEmail("u3@example.com")
	u4 := models.NewDBUsersFixture().WithUsername("u4").WithEmail("u4@example.com")

	mWin := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Termine).
		WithCreatorId(u1.Id)
	mLoss := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Termine).
		WithCreatorId(u1.Id)
	mDraw := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Termine).
		WithCreatorId(u1.Id)
	mNotFinished := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.EnCours).
		WithCreatorId(u1.Id)
	mNoScore := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Termine).
		WithCreatorId(u1.Id)

	mWin.Score1 = ptr(3)
	mWin.Score2 = ptr(1)
	mLoss.Score1 = ptr(2)
	mLoss.Score2 = ptr(4)
	mDraw.Score1 = ptr(2)
	mDraw.Score2 = ptr(2)
	mNotFinished.Score1 = ptr(5)
	mNotFinished.Score2 = ptr(0)
	mNoScore.Score1 = nil
	mNoScore.Score2 = nil

	fixturesCommon := DBFixtures{
		Courts: []models.DBCourt{court},
		Matches: []models.DBMatches{
			mWin, mLoss, mDraw, mNotFinished, mNoScore,
		},
		Users: []models.DBUsers{u1, u2, u3, u4},
		UserMatches: []models.DBUserMatch{
			{UserID: u1.Id, MatchID: mWin.Id, Team: 1, CreatedAt: time.Now()},
			{UserID: u1.Id, MatchID: mLoss.Id, Team: 1, CreatedAt: time.Now()},
			{UserID: u1.Id, MatchID: mDraw.Id, Team: 1, CreatedAt: time.Now()},
			{UserID: u1.Id, MatchID: mNotFinished.Id, Team: 1, CreatedAt: time.Now()},
			{UserID: u1.Id, MatchID: mNoScore.Id, Team: 1, CreatedAt: time.Now()},
			{UserID: u2.Id, MatchID: mLoss.Id, Team: 2, CreatedAt: time.Now()},
			{UserID: u3.Id, MatchID: mDraw.Id, Team: 2, CreatedAt: time.Now()},
		},
	}

	testCases := []testCase{
		{
			name:     "u1 -> 1 win / 2 admissibles = 50%",
			fixtures: fixturesCommon,
			userID:   u1.Id,
			expected: expected{wantNil: false, value: 50},
		},
		{
			name:     "u2 -> 1 win / 1 admissible = 100%",
			fixtures: fixturesCommon,
			userID:   u2.Id,
			expected: expected{wantNil: false, value: 100},
		},
		{
			name:     "u3 -> seulement des nuls => nil",
			fixtures: fixturesCommon,
			userID:   u3.Id,
			expected: expected{wantNil: true},
		},
		{
			name:     "u4 -> aucun match => nil",
			fixtures: fixturesCommon,
			userID:   u4.Id,
			expected: expected{wantNil: true},
		},
		{
			name:     "inconnu -> nil",
			fixtures: fixturesCommon,
			userID:   uuid.NewString(),
			expected: expected{wantNil: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()

			s.loadFixtures(tc.fixtures)

			ctx := context.Background()
			got, err := s.db.GetUserWinrate(ctx, tc.userID)
			require.NoError(t, err)

			if tc.expected.wantNil {
				require.Nil(t, got, "winrate should be nil")
			} else {
				require.NotNil(t, got, "winrate should not be nil")
				require.Equal(t, tc.expected.value, *got)
			}
		})
	}
}

func TestDatabase_GetUsersByMatchIDs(t *testing.T) {
	type expected struct {
		matchUsers map[string][]string
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		matchIDs []string
		expected expected
	}

	court := models.NewDBCourtFixture()
	user1 := models.NewDBUsersFixture().WithUsername("Alice").WithEmail("email1@gmail.com")
	user2 := models.NewDBUsersFixture().WithUsername("Bob").WithEmail("email2@gmail.com")
	user3 := models.NewDBUsersFixture().WithUsername("Charlie").WithEmail("email3@gmail.com")
	match1 := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCreatorId(user1.Id)
	match2 := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCreatorId(user1.Id)

	testCases := []testCase{
		{
			name: "Deux matchs avec des users",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match1, match2},
				Users:   []models.DBUsers{user1, user2, user3},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(user1.Id).WithMatchId(match1.Id),
					models.NewDBUserMatchFixture().WithUserId(user2.Id).WithMatchId(match1.Id),
					models.NewDBUserMatchFixture().WithUserId(user3.Id).WithMatchId(match2.Id),
				},
			},
			matchIDs: []string{match1.Id, match2.Id},
			expected: expected{
				matchUsers: map[string][]string{
					match1.Id: {"Alice", "Bob"},
					match2.Id: {"Charlie"},
				},
			},
		},
		{
			name: "Aucun user pour les matchs demandés",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{user1},
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match1, match2},
			},
			matchIDs: []string{match1.Id, match2.Id},
			expected: expected{
				matchUsers: map[string][]string{},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()
			got, err := s.db.GetUsersByMatchIDs(ctx, c.matchIDs)
			require.NoError(t, err)

			for mid, usernames := range c.expected.matchUsers {
				require.Contains(t, got, mid)
				gotUsernames := make([]string, len(got[mid]))
				for i, u := range got[mid] {
					gotUsernames[i] = u.Username
				}
				require.ElementsMatch(t, usernames, gotUsernames)
			}
		})
	}
}

func TestDatabase_GetUsersByMatchId(t *testing.T) {
	type expected struct {
		userIDs []string
		found   bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		matchID  string
		expected expected
	}

	court := models.NewDBCourtFixture()
	user1 := models.NewDBUsersFixture().WithUsername("alice").WithEmail("alice@test.com")
	user2 := models.NewDBUsersFixture().WithUsername("bob").WithEmail("bob@test.com")
	match := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCreatorId(user1.Id)

	testCases := []testCase{
		{
			name: "Match with 2 users",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
				Users:   []models.DBUsers{user1, user2},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(user1.Id).WithMatchId(match.Id),
					models.NewDBUserMatchFixture().WithUserId(user2.Id).WithMatchId(match.Id),
				},
			},
			matchID: match.Id,
			expected: expected{
				userIDs: []string{user1.Id, user2.Id},
				found:   true,
			},
		},
		{
			name: "Match exists but no users",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{user1},
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
			},
			matchID: match.Id,
			expected: expected{
				userIDs: []string{},
				found:   false,
			},
		},
		{
			name:     "Unknown match",
			fixtures: DBFixtures{},
			matchID:  uuid.NewString(),
			expected: expected{
				userIDs: []string{},
				found:   false,
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

			users, err := s.db.GetUsersByMatchId(ctx, tc.matchID)
			require.NoError(t, err)

			if !tc.expected.found {
				require.Empty(t, users)
				return
			}

			require.Len(t, users, len(tc.expected.userIDs))
			gotIDs := make([]string, len(users))
			for i, u := range users {
				gotIDs[i] = u.Id
			}
			require.ElementsMatch(t, tc.expected.userIDs, gotIDs)
		})
	}
}

func TestDatabase_DeleteMatch(t *testing.T) {
	type expected struct {
		shouldExist bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		matchID  string
		expected expected
	}

	court := models.NewDBCourtFixture()
	user := models.NewDBUsersFixture()
	match := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCreatorId(user.Id)

	testCases := []testCase{
		{
			name: "Delete existing match",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{user},
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match},
			},
			matchID: match.Id,
			expected: expected{
				shouldExist: false,
			},
		},
		{
			name: "Delete non-existing match (no error)",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court},
			},
			matchID: uuid.NewString(),
			expected: expected{
				shouldExist: false,
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

			err := s.db.DeleteMatch(ctx, tc.matchID)
			require.NoError(t, err)

			got, err := s.db.GetMatchById(ctx, tc.matchID)
			require.NoError(t, err)

			if tc.expected.shouldExist {
				require.NotNil(t, got, "Match devrait encore exister")
			} else {
				require.Nil(t, got, "Match ne devrait plus exister")
			}
		})
	}
}

func TestDatabase_CountUsersByMatch(t *testing.T) {
	type expected struct {
		count int
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		matchID  string
		expected expected
	}

	court := models.NewDBCourtFixture()
	user1 := models.NewDBUsersFixture().WithUsername("alice").WithEmail("alice@test.com")
	user2 := models.NewDBUsersFixture().WithUsername("bob").WithEmail("bob@test.com")
	match1 := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCreatorId(user1.Id)
	match2 := models.NewDBMatchesFixture().WithCourtId(court.Id).WithCreatorId(user1.Id)

	testCases := []testCase{
		{
			name: "Match with 2 users",
			fixtures: DBFixtures{
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match1, match2},
				Users:   []models.DBUsers{user1, user2},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().WithUserId(user1.Id).WithMatchId(match1.Id),
					models.NewDBUserMatchFixture().WithUserId(user2.Id).WithMatchId(match1.Id),
				},
			},
			matchID: match1.Id,
			expected: expected{
				count: 2,
			},
		},
		{
			name: "Match exists but has no users",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{user1},
				Courts:  []models.DBCourt{court},
				Matches: []models.DBMatches{match2},
			},
			matchID: match2.Id,
			expected: expected{
				count: 0,
			},
		},
		{
			name:     "Match does not exist",
			fixtures: DBFixtures{},
			matchID:  uuid.NewString(),
			expected: expected{
				count: 0,
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

			got, err := s.db.CountUsersByMatch(ctx, tc.matchID)
			require.NoError(t, err)
			require.Equal(t, tc.expected.count, got)
		})
	}
}
