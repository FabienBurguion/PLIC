package database

import (
	"PLIC/models"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDatabase_GetMatchScoreVote(t *testing.T) {
	type expected struct {
		shouldExist bool
		team        int
		score1      int
		score2      int
	}

	type testCase struct {
		name        string
		fixtures    DBFixtures
		insertVotes []models.DBMatchScoreVote
		queryMatch  string
		queryUser   string
		want        expected
	}

	u1 := models.NewDBUsersFixture()
	u2 := models.NewDBUsersFixture().
		WithUsername("user2").
		WithEmail("email2")
	c := models.NewDBCourtFixture()
	m := models.NewDBMatchesFixture().
		WithCourtId(c.Id).
		WithCurrentState(models.ManqueScore)

	tests := []testCase{
		{
			name: "Vote exists",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			insertVotes: []models.DBMatchScoreVote{
				{MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 3, Score2: 2},
			},
			queryMatch: m.Id,
			queryUser:  u1.Id,
			want: expected{
				shouldExist: true,
				team:        1,
				score1:      3,
				score2:      2,
			},
		},
		{
			name: "No vote for this user",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1, u2},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			insertVotes: []models.DBMatchScoreVote{
				{MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 2, Score2: 2},
			},
			queryMatch: m.Id,
			queryUser:  u2.Id,
			want: expected{
				shouldExist: false,
			},
		},
		{
			name: "Unknown match id",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			insertVotes: []models.DBMatchScoreVote{
				{MatchId: m.Id, UserId: u1.Id, Team: 2, Score1: 1, Score2: 0},
			},
			queryMatch: uuid.NewString(),
			queryUser:  u1.Id,
			want: expected{
				shouldExist: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(tc.fixtures)

			ctx := context.Background()

			for _, v := range tc.insertVotes {
				err := s.db.UpsertMatchScoreVote(ctx, v)
				require.NoError(t, err)
			}

			got, err := s.db.GetMatchScoreVote(ctx, tc.queryMatch, tc.queryUser)
			require.NoError(t, err)

			if !tc.want.shouldExist {
				require.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			require.Equal(t, tc.queryMatch, got.MatchId)
			require.Equal(t, tc.queryUser, got.UserId)
			require.Equal(t, tc.want.team, got.Team)
			require.Equal(t, tc.want.score1, got.Score1)
			require.Equal(t, tc.want.score2, got.Score2)
		})
	}
}

func TestDatabase_UpsertMatchScoreVote(t *testing.T) {
	type expected struct {
		team        int
		score1      int
		score2      int
		shouldExist bool
	}

	type testCase struct {
		name       string
		fixtures   DBFixtures
		firstVote  models.DBMatchScoreVote
		updateVote *models.DBMatchScoreVote
		checks     []struct {
			matchID string
			userID  string
			want    expected
		}
	}

	u1 := models.NewDBUsersFixture()
	u2 := models.NewDBUsersFixture().
		WithUsername("user2").
		WithEmail("email2")
	c := models.NewDBCourtFixture()
	m := models.NewDBMatchesFixture().
		WithCourtId(c.Id).
		WithCurrentState(models.ManqueScore)

	tests := []testCase{
		{
			name: "Insert new vote",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			firstVote: models.DBMatchScoreVote{
				MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 3, Score2: 2,
			},
			checks: []struct {
				matchID string
				userID  string
				want    expected
			}{
				{m.Id, u1.Id, expected{team: 1, score1: 3, score2: 2, shouldExist: true}},
			},
		},
		{
			name: "Update existing vote via conflict",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			firstVote: models.DBMatchScoreVote{
				MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 1, Score2: 0,
			},
			updateVote: &models.DBMatchScoreVote{
				MatchId: m.Id, UserId: u1.Id, Team: 2, Score1: 5, Score2: 4,
			},
			checks: []struct {
				matchID string
				userID  string
				want    expected
			}{
				{m.Id, u1.Id, expected{team: 2, score1: 5, score2: 4, shouldExist: true}},
			},
		},
		{
			name: "Two users vote on same match",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1, u2},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			firstVote: models.DBMatchScoreVote{
				MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 2, Score2: 2,
			},
			updateVote: &models.DBMatchScoreVote{
				MatchId: m.Id, UserId: u2.Id, Team: 2, Score1: 2, Score2: 2,
			},
			checks: []struct {
				matchID string
				userID  string
				want    expected
			}{
				{m.Id, u1.Id, expected{team: 1, score1: 2, score2: 2, shouldExist: true}},
				{m.Id, u2.Id, expected{team: 2, score1: 2, score2: 2, shouldExist: true}},
			},
		},
		{
			name: "Vote not found",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			firstVote: models.DBMatchScoreVote{
				MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 1, Score2: 1,
			},
			checks: []struct {
				matchID string
				userID  string
				want    expected
			}{
				{m.Id, u2.Id, expected{shouldExist: false}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() {
				if err := cleanup(); err != nil {
					t.Logf("cleanup error: %v", err)
				}
			}()
			s.loadFixtures(tc.fixtures)

			ctx := context.Background()

			err := s.db.UpsertMatchScoreVote(ctx, tc.firstVote)
			require.NoError(t, err)

			if tc.updateVote != nil {
				err = s.db.UpsertMatchScoreVote(ctx, *tc.updateVote)
				require.NoError(t, err)
			}

			for _, chk := range tc.checks {
				got, err := s.db.GetMatchScoreVote(ctx, chk.matchID, chk.userID)
				if !chk.want.shouldExist {
					require.NoError(t, err)
					require.Nil(t, got)
					continue
				}
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Equal(t, chk.matchID, got.MatchId)
				require.Equal(t, chk.userID, got.UserId)
				require.Equal(t, chk.want.team, got.Team)
				require.Equal(t, chk.want.score1, got.Score1)
				require.Equal(t, chk.want.score2, got.Score2)
			}
		})
	}
}

func TestDatabase_HasConsensusScore(t *testing.T) {
	type expected struct {
		consensus bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		votes    []models.DBMatchScoreVote
		callArgs struct {
			matchID string
			team    int
			s1, s2  int
		}
		want expected
	}

	u1 := models.NewDBUsersFixture().WithUsername("user1").WithEmail("email1")
	u2 := models.NewDBUsersFixture().WithUsername("user2").WithEmail("email2")
	u3 := models.NewDBUsersFixture().WithUsername("user3").WithEmail("email3")
	c := models.NewDBCourtFixture()
	m := models.NewDBMatchesFixture().
		WithCourtId(c.Id).
		WithCurrentState(models.ManqueScore)

	testCases := []testCase{
		{
			name: "No votes => no consensus",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1, u2},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			votes: nil,
			callArgs: struct {
				matchID string
				team    int
				s1, s2  int
			}{m.Id, 1, 3, 2},
			want: expected{consensus: false},
		},
		{
			name: "Only same-team vote => no consensus",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1, u2},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			votes: []models.DBMatchScoreVote{
				{MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 3, Score2: 2},
			},
			callArgs: struct {
				matchID string
				team    int
				s1, s2  int
			}{m.Id, 1, 3, 2},
			want: expected{consensus: false},
		},
		{
			name: "Other team different score => no consensus",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1, u2},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			votes: []models.DBMatchScoreVote{
				{MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 3, Score2: 2},
				{MatchId: m.Id, UserId: u2.Id, Team: 2, Score1: 4, Score2: 4},
			},
			callArgs: struct {
				matchID string
				team    int
				s1, s2  int
			}{m.Id, 1, 3, 2},
			want: expected{consensus: false},
		},
		{
			name: "Other team same score => consensus",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1, u2, u3},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			votes: []models.DBMatchScoreVote{
				{MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 3, Score2: 2},
				{MatchId: m.Id, UserId: u2.Id, Team: 2, Score1: 3, Score2: 2},
			},
			callArgs: struct {
				matchID string
				team    int
				s1, s2  int
			}{m.Id, 1, 3, 2},
			want: expected{consensus: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
			s.loadFixtures(tc.fixtures)

			ctx := context.Background()

			for _, v := range tc.votes {
				err := s.db.UpsertMatchScoreVote(ctx, v)
				require.NoError(t, err)
			}

			got, err := s.db.HasConsensusScore(ctx, tc.callArgs.matchID, tc.callArgs.team, tc.callArgs.s1, tc.callArgs.s2)
			require.NoError(t, err)
			require.Equal(t, tc.want.consensus, got)
		})
	}
}

func TestDatabase_HasOtherTeamVote(t *testing.T) {
	type expected struct {
		hasOther bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		votes    []models.DBMatchScoreVote
		callArgs struct {
			matchID string
			team    int
			userID  string
		}
		want expected
	}

	u1 := models.NewDBUsersFixture().WithUsername("user1").WithEmail("email1")
	u2 := models.NewDBUsersFixture().WithUsername("user2").WithEmail("email2")
	u3 := models.NewDBUsersFixture().WithUsername("user3").WithEmail("email3")
	c := models.NewDBCourtFixture()
	m := models.NewDBMatchesFixture().
		WithCourtId(c.Id).
		WithCurrentState(models.ManqueScore)

	tests := []testCase{
		{
			name: "No votes => false",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1, u2, u3},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			votes: nil,
			callArgs: struct {
				matchID string
				team    int
				userID  string
			}{m.Id, 1, u1.Id},
			want: expected{hasOther: false},
		},
		{
			name: "Same user vote only => false",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			votes: []models.DBMatchScoreVote{
				{MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 1, Score2: 0},
			},
			callArgs: struct {
				matchID string
				team    int
				userID  string
			}{m.Id, 1, u1.Id},
			want: expected{hasOther: false},
		},
		{
			name: "Another user same team => true",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1, u3},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			votes: []models.DBMatchScoreVote{
				{MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 2, Score2: 2},
				{MatchId: m.Id, UserId: u3.Id, Team: 1, Score1: 3, Score2: 3},
			},
			callArgs: struct {
				matchID string
				team    int
				userID  string
			}{m.Id, 1, u1.Id},
			want: expected{hasOther: true},
		},
		{
			name: "Another user other team => false",
			fixtures: DBFixtures{
				Users:   []models.DBUsers{u1, u2},
				Courts:  []models.DBCourt{c},
				Matches: []models.DBMatches{m},
			},
			votes: []models.DBMatchScoreVote{
				{MatchId: m.Id, UserId: u1.Id, Team: 1, Score1: 4, Score2: 4},
				{MatchId: m.Id, UserId: u2.Id, Team: 2, Score1: 4, Score2: 4},
			},
			callArgs: struct {
				matchID string
				team    int
				userID  string
			}{m.Id, 1, u1.Id},
			want: expected{hasOther: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()
			s.loadFixtures(tc.fixtures)

			ctx := context.Background()

			for _, v := range tc.votes {
				err := s.db.UpsertMatchScoreVote(ctx, v)
				require.NoError(t, err)
			}

			got, err := s.db.HasOtherTeamVote(ctx, tc.callArgs.matchID, tc.callArgs.team, tc.callArgs.userID)
			require.NoError(t, err)
			require.Equal(t, tc.want.hasOther, got)
		})
	}
}
