package database

import (
	"PLIC/models"
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDatabase_GetRankedFieldsByUserID(t *testing.T) {
	type expected struct {
		fields  []models.Field
		isError bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		param    string
		expected expected
	}

	userID := uuid.NewString()
	otherUser1 := uuid.NewString()
	otherUser2 := uuid.NewString()

	courtID1 := uuid.NewString()
	courtID2 := uuid.NewString()
	courtID3 := uuid.NewString()

	matchID1 := uuid.NewString()
	matchID2 := uuid.NewString()
	matchID3 := uuid.NewString()

	testCases := []testCase{
		{
			name: "User has rankings on multiple courts with different local ranks (per sport)",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userID).
						WithUsername("user1").
						WithEmail("email1"),
					models.NewDBUsersFixture().
						WithId(otherUser1).
						WithUsername("user2").
						WithEmail("email2"),
					models.NewDBUsersFixture().
						WithId(otherUser2).
						WithUsername("user3").
						WithEmail("email3"),
				},
				Courts: []models.DBCourt{
					models.NewDBCourtFixture().
						WithId(courtID1).
						WithName("Central Park"),
					models.NewDBCourtFixture().
						WithId(courtID2).
						WithName("Stade de Lyon"),
					models.NewDBCourtFixture().
						WithId(courtID3).
						WithName("Playground"),
				},
				Rankings: []models.DBRanking{
					models.NewDBRankingFixture().
						WithUserId(userID).
						WithCourtId(courtID1).
						WithSport(models.Basket).
						WithElo(1200),
					models.NewDBRankingFixture().
						WithUserId(otherUser1).
						WithCourtId(courtID1).
						WithSport(models.Basket).
						WithElo(1300),
					models.NewDBRankingFixture().
						WithUserId(otherUser2).
						WithCourtId(courtID1).
						WithSport(models.Basket).
						WithElo(1250),

					models.NewDBRankingFixture().
						WithUserId(userID).
						WithCourtId(courtID2).
						WithSport(models.Foot).
						WithElo(1300),
					models.NewDBRankingFixture().
						WithUserId(otherUser1).
						WithCourtId(courtID2).
						WithSport(models.Foot).
						WithElo(1000),

					models.NewDBRankingFixture().
						WithUserId(userID).
						WithCourtId(courtID3).
						WithSport(models.PingPong).
						WithElo(1250),
					models.NewDBRankingFixture().
						WithUserId(otherUser1).
						WithCourtId(courtID3).
						WithSport(models.PingPong).
						WithElo(1250),
					models.NewDBRankingFixture().
						WithUserId(otherUser2).
						WithCourtId(courtID3).
						WithSport(models.PingPong).
						WithElo(1100),
				},
				Matches: []models.DBMatches{
					models.NewDBMatchesFixture().
						WithId(matchID1).
						WithCourtId(courtID1).
						WithSport(models.Basket).
						WithCreatorId(userID),
					models.NewDBMatchesFixture().
						WithId(matchID2).
						WithCourtId(courtID2).
						WithSport(models.Foot).
						WithCreatorId(userID),
					models.NewDBMatchesFixture().
						WithId(matchID3).
						WithCourtId(courtID3).
						WithSport(models.PingPong).
						WithCreatorId(userID),
				},
				UserMatches: []models.DBUserMatch{
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID1).
						WithTeam(1),
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID2).
						WithTeam(1),
					models.NewDBUserMatchFixture().
						WithUserId(userID).
						WithMatchId(matchID3).
						WithTeam(2),
				},
			},
			param: userID,
			expected: expected{
				fields: []models.Field{
					models.NewFieldFixture().
						WithRanking(3).
						WithName("Central Park").
						WithScore(1200).
						WithSport(models.Basket),
					models.NewFieldFixture().
						WithRanking(1).
						WithName("Playground").
						WithScore(1250).
						WithSport(models.PingPong),
					models.NewFieldFixture().
						WithRanking(1).
						WithName("Stade de Lyon").
						WithScore(1300).
						WithSport(models.Foot),
				},
				isError: false,
			},
		},
		{
			name: "User has no ranked courts",
			fixtures: DBFixtures{
				Users: []models.DBUsers{
					models.NewDBUsersFixture().
						WithId(userID),
				},
			},
			param: userID,
			expected: expected{
				fields:  []models.Field{},
				isError: false,
			},
		},
		{
			name:     "Unknown user",
			fixtures: DBFixtures{},
			param:    uuid.NewString(),
			expected: expected{
				fields:  []models.Field{},
				isError: false,
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
			fields, err := s.db.GetRankedFieldsByUserID(ctx, c.param)

			if c.expected.isError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(c.expected.fields), len(fields))

			sort.Slice(fields, func(i, j int) bool {
				if fields[i].Name == fields[j].Name {
					return string(fields[i].Sport) < string(fields[j].Sport)
				}
				return fields[i].Name < fields[j].Name
			})
			sort.Slice(c.expected.fields, func(i, j int) bool {
				if c.expected.fields[i].Name == c.expected.fields[j].Name {
					return string(c.expected.fields[i].Sport) < string(c.expected.fields[j].Sport)
				}
				return c.expected.fields[i].Name < c.expected.fields[j].Name
			})

			for i := range fields {
				require.Equal(t, c.expected.fields[i].Ranking, fields[i].Ranking, "wrong ranking for %s (%s)", fields[i].Name, fields[i].Sport)
				require.Equal(t, c.expected.fields[i].Name, fields[i].Name)
				require.Equal(t, c.expected.fields[i].Elo, fields[i].Elo)
				require.Equal(t, c.expected.fields[i].Sport, fields[i].Sport)
			}
		})
	}
}

func TestDatabase_InsertRanking(t *testing.T) {
	type testCase struct {
		name      string
		fixtures  DBFixtures
		param     models.DBRanking
		expected  int
		preInsert bool
	}

	user := models.NewDBUsersFixture()
	court := models.NewDBCourtFixture()

	sport := models.Basket

	initialRanking := models.NewDBRankingFixture().
		WithUserId(user.Id).
		WithCourtId(court.Id).
		WithSport(sport).
		WithElo(1450)

	updatedRanking := initialRanking
	updatedRanking.Elo = 1600
	updatedRanking.UpdatedAt = time.Now()

	testCases := []testCase{
		{
			name: "Insert new ranking",
			fixtures: DBFixtures{
				Users:  []models.DBUsers{user},
				Courts: []models.DBCourt{court},
			},
			param:    initialRanking,
			expected: 1450,
		},
		{
			name: "Update existing ranking on conflict (user,court,sport)",
			fixtures: DBFixtures{
				Users:  []models.DBUsers{user},
				Courts: []models.DBCourt{court},
			},
			param:     updatedRanking,
			expected:  1600,
			preInsert: true,
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

			if c.preInsert {
				err := s.db.InsertRanking(ctx, initialRanking)
				require.NoError(t, err)
			}

			err := s.db.InsertRanking(ctx, c.param)
			require.NoError(t, err)

			ranking, err := s.db.GetRankingByUserCourtSport(ctx, c.param.UserID, c.param.CourtID, c.param.Sport)
			require.NoError(t, err)
			require.NotNil(t, ranking)
			require.Equal(t, c.expected, ranking.Elo)
		})
	}
}

func TestDatabase_GetRankingByUserAndCourt(t *testing.T) {
	type expected struct {
		found bool
		elo   int
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		userID   string
		courtID  string
		expected expected
	}

	user := models.NewDBUsersFixture()
	court := models.NewDBCourtFixture()

	ranking := models.NewDBRankingFixture().
		WithUserId(user.Id).
		WithCourtId(court.Id).
		WithElo(1500)

	testCases := []testCase{
		{
			name: "Ranking exists",
			fixtures: DBFixtures{
				Users:    []models.DBUsers{user},
				Courts:   []models.DBCourt{court},
				Rankings: []models.DBRanking{ranking},
			},
			userID:  user.Id,
			courtID: court.Id,
			expected: expected{
				found: true,
				elo:   1500,
			},
		},
		{
			name: "Ranking not found",
			fixtures: DBFixtures{
				Users:  []models.DBUsers{user},
				Courts: []models.DBCourt{court},
			},
			userID:  user.Id,
			courtID: court.Id,
			expected: expected{
				found: false,
			},
		},
		{
			name: "Wrong court id",
			fixtures: DBFixtures{
				Users:    []models.DBUsers{user},
				Courts:   []models.DBCourt{court},
				Rankings: []models.DBRanking{ranking},
			},
			userID:  user.Id,
			courtID: uuid.NewString(),
			expected: expected{
				found: false,
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
			got, err := s.db.GetRankingByUserAndCourt(ctx, c.userID, c.courtID)

			require.NoError(t, err)

			if !c.expected.found {
				require.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			require.Equal(t, c.userID, got.UserID)
			require.Equal(t, c.courtID, got.CourtID)
			require.Equal(t, c.expected.elo, got.Elo)
		})
	}
}

func TestDatabase_GetRankingsByCourtID(t *testing.T) {
	type expected struct {
		wantLen   int
		checkSort bool
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		courtID  string
		sport    models.Sport
		expected expected
	}

	// ---------- Fixtures communes ----------
	courtA := models.NewDBCourtFixture()
	courtB := models.NewDBCourtFixture()

	u1 := models.NewDBUsersFixture().WithUsername("alice").WithEmail("alice@example.com")
	u2 := models.NewDBUsersFixture().WithUsername("bob").WithEmail("bob@example.com")
	u3 := models.NewDBUsersFixture().WithUsername("carol").WithEmail("carol@example.com")
	u4 := models.NewDBUsersFixture().WithUsername("dave").WithEmail("dave@example.com")

	now := time.Now()

	// Court A - PingPong
	rA1 := models.NewDBRankingFixture().
		WithUserId(u1.Id).
		WithCourtId(courtA.Id).
		WithSport(models.PingPong).
		WithElo(1200)

	rA2 := models.NewDBRankingFixture().
		WithUserId(u2.Id).
		WithCourtId(courtA.Id).
		WithSport(models.PingPong).
		WithElo(1100)

	rA3 := models.NewDBRankingFixture().
		WithUserId(u3.Id).
		WithCourtId(courtA.Id).
		WithSport(models.PingPong).
		WithElo(1100)

	// Court A - autre sport (doit être ignoré)
	rAOtherSport := models.NewDBRankingFixture().
		WithUserId(u4.Id).
		WithCourtId(courtA.Id).
		WithSport(models.Foot).
		WithElo(2000)

	// Court B - PingPong
	rB1 := models.NewDBRankingFixture().
		WithUserId(u4.Id).
		WithCourtId(courtB.Id).
		WithSport(models.PingPong).
		WithElo(1500)

	for _, r := range []*models.DBRanking{
		&rA1, &rA2, &rA3, &rAOtherSport, &rB1,
	} {
		r.CreatedAt = now
		r.UpdatedAt = now
	}

	// ---------- Test cases ----------
	testCases := []testCase{
		{
			name: "Nominal - filtre par court et sport + tri elo puis user_id",
			fixtures: DBFixtures{
				Courts:   []models.DBCourt{courtA},
				Users:    []models.DBUsers{u1, u2, u3, u4},
				Rankings: []models.DBRanking{rA1, rA2, rA3, rAOtherSport},
			},
			courtID: courtA.Id,
			sport:   models.PingPong,
			expected: expected{
				wantLen:   3,
				checkSort: true,
			},
		},
		{
			name: "Aucun ranking pour ce sport",
			fixtures: DBFixtures{
				Courts:   []models.DBCourt{courtA},
				Users:    []models.DBUsers{u1, u2, u3},
				Rankings: []models.DBRanking{rA1, rA2, rA3},
			},
			courtID: courtA.Id,
			sport:   models.Basket,
			expected: expected{
				wantLen:   0,
				checkSort: false,
			},
		},
		{
			name: "Plusieurs courts - filtre uniquement sur le bon court",
			fixtures: DBFixtures{
				Courts:   []models.DBCourt{courtA, courtB},
				Users:    []models.DBUsers{u1, u2, u3, u4},
				Rankings: []models.DBRanking{rA1, rA2, rA3, rB1},
			},
			courtID: courtB.Id,
			sport:   models.PingPong,
			expected: expected{
				wantLen:   1,
				checkSort: true,
			},
		},
		{
			name: "Court inexistant - retourne liste vide",
			fixtures: DBFixtures{
				Courts:   []models.DBCourt{courtA},
				Users:    []models.DBUsers{u1, u2, u3},
				Rankings: []models.DBRanking{rA1, rA2, rA3},
			},
			courtID: uuid.NewString(),
			sport:   models.PingPong,
			expected: expected{
				wantLen:   0,
				checkSort: false,
			},
		},
	}

	// ---------- Helper de tri ----------
	isSorted := func(rs []models.DBRanking) bool {
		return sort.SliceIsSorted(rs, func(i, j int) bool {
			if rs[i].Elo == rs[j].Elo {
				return rs[i].UserID < rs[j].UserID
			}
			return rs[i].Elo > rs[j].Elo
		})
	}

	// ---------- Exécution ----------
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Service{}
			cleanup := s.InitServiceTest()
			defer func() { _ = cleanup() }()

			s.loadFixtures(tc.fixtures)

			out, err := s.db.GetRankingsByCourtID(
				context.Background(),
				tc.courtID,
				tc.sport,
			)

			require.NoError(t, err)
			require.Len(t, out, tc.expected.wantLen)

			if tc.expected.checkSort {
				require.True(
					t,
					isSorted(out),
					"les rankings doivent être triés par (elo asc, user_id asc)",
				)
			}
		})
	}
}

func TestDatabase_GetRankingByUserCourtSport(t *testing.T) {
	type expected struct {
		found bool
		elo   int
	}

	type testCase struct {
		name     string
		fixtures DBFixtures
		userID   string
		courtID  string
		sport    models.Sport
		expected expected
	}

	user := models.NewDBUsersFixture()
	court := models.NewDBCourtFixture()

	ranking := models.NewDBRankingFixture().
		WithUserId(user.Id).
		WithCourtId(court.Id).
		WithSport(models.Foot).
		WithElo(1500)

	testCases := []testCase{
		{
			name: "Ranking exists for (user,court,sport)",
			fixtures: DBFixtures{
				Users:    []models.DBUsers{user},
				Courts:   []models.DBCourt{court},
				Rankings: []models.DBRanking{ranking},
			},
			userID:  user.Id,
			courtID: court.Id,
			sport:   models.Foot,
			expected: expected{
				found: true,
				elo:   1500,
			},
		},
		{
			name: "Ranking not found (no row)",
			fixtures: DBFixtures{
				Users:  []models.DBUsers{user},
				Courts: []models.DBCourt{court},
			},
			userID:  user.Id,
			courtID: court.Id,
			sport:   models.Foot,
			expected: expected{
				found: false,
			},
		},
		{
			name: "Wrong court id",
			fixtures: DBFixtures{
				Users:    []models.DBUsers{user},
				Courts:   []models.DBCourt{court},
				Rankings: []models.DBRanking{ranking},
			},
			userID:  user.Id,
			courtID: uuid.NewString(),
			sport:   models.Foot,
			expected: expected{
				found: false,
			},
		},
		{
			name: "Wrong sport",
			fixtures: DBFixtures{
				Users:    []models.DBUsers{user},
				Courts:   []models.DBCourt{court},
				Rankings: []models.DBRanking{ranking},
			},
			userID:  user.Id,
			courtID: court.Id,
			sport:   models.Basket,
			expected: expected{
				found: false,
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
			got, err := s.db.GetRankingByUserCourtSport(ctx, c.userID, c.courtID, c.sport)

			require.NoError(t, err)

			if !c.expected.found {
				require.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			require.Equal(t, c.userID, got.UserID)
			require.Equal(t, c.courtID, got.CourtID)
			require.Equal(t, c.sport, got.Sport)
			require.Equal(t, c.expected.elo, got.Elo)
		})
	}
}
