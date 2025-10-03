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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
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
	match1 := models.NewDBMatchesFixture().WithCourtId(court.Id)
	match2 := models.NewDBMatchesFixture().WithCourtId(court.Id)

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

	match1 := models.NewDBMatchesFixture().
		WithCourtId(court.Id)
	match2 := models.NewDBMatchesFixture().
		WithCourtId(court.Id)

	u1 := models.NewDBUsersFixture().WithUsername("u1").WithEmail("u1@example.com")
	u2 := models.NewDBUsersFixture().WithUsername("u2").WithEmail("u2@example.com")
	u3 := models.NewDBUsersFixture().WithUsername("u3").WithEmail("u3@example.com")

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

	mWin := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Termine)
	mLoss := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Termine)
	mDraw := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Termine)
	mNotFinished := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.EnCours)
	mNoScore := models.NewDBMatchesFixture().
		WithCourtId(court.Id).
		WithCurrentState(models.Termine)

	u1 := models.NewDBUsersFixture().WithUsername("u1").WithEmail("u1@example.com")
	u2 := models.NewDBUsersFixture().WithUsername("u2").WithEmail("u2@example.com")
	u3 := models.NewDBUsersFixture().WithUsername("u3").WithEmail("u3@example.com")
	u4 := models.NewDBUsersFixture().WithUsername("u4").WithEmail("u4@example.com")

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
	match1 := models.NewDBMatchesFixture().WithCourtId(court.Id)
	match2 := models.NewDBMatchesFixture().WithCourtId(court.Id)
	user1 := models.NewDBUsersFixture().WithUsername("Alice").WithEmail("email1@gmail.com")
	user2 := models.NewDBUsersFixture().WithUsername("Bob").WithEmail("email2@gmail.com")
	user3 := models.NewDBUsersFixture().WithUsername("Charlie").WithEmail("email3@gmail.com")

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

func TestDatabase_GetCourtsByIDs(t *testing.T) {
	type expected struct {
		courts map[string]string
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		ids      []string
		expected expected
	}

	court1 := models.NewDBCourtFixture().WithName("Court A")
	court2 := models.NewDBCourtFixture().WithName("Court B")

	testCases := []testCase{
		{
			name: "Deux courts existants",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{court1, court2},
			},
			ids: []string{court1.Id, court2.Id},
			expected: expected{
				courts: map[string]string{
					court1.Id: "Court A",
					court2.Id: "Court B",
				},
			},
		},
		{
			name: "Aucun court trouvé",
			fixtures: DBFixtures{
				Courts: []models.DBCourt{},
			},
			ids:      []string{"nonexistent-id"},
			expected: expected{courts: map[string]string{}},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
			s.loadFixtures(c.fixtures)

			ctx := context.Background()
			got, err := s.db.GetCourtsByIDs(ctx, c.ids)
			require.NoError(t, err)

			gotMap := make(map[string]string)
			for _, ct := range got {
				gotMap[ct.Id] = ct.Name
			}

			require.Equal(t, c.expected.courts, gotMap)
		})
	}
}
