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

	testCases := []testCase{
		{
			name: "User has rankings on multiple courts with different local ranks",
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
					// Court 1
					models.NewDBRankingFixture().
						WithUserId(userID).
						WithCourtId(courtID1).
						WithElo(1200),
					models.NewDBRankingFixture().
						WithUserId(otherUser1).
						WithCourtId(courtID1).
						WithElo(1300),
					models.NewDBRankingFixture().
						WithUserId(otherUser2).
						WithCourtId(courtID1).
						WithElo(1250),

					// Court 2
					models.NewDBRankingFixture().
						WithUserId(userID).
						WithCourtId(courtID2).
						WithElo(1300),
					models.NewDBRankingFixture().
						WithUserId(otherUser1).
						WithCourtId(courtID2).
						WithElo(1000),

					// Court 3
					models.NewDBRankingFixture().
						WithUserId(userID).
						WithCourtId(courtID3).
						WithElo(1250),
					models.NewDBRankingFixture().
						WithUserId(otherUser1).
						WithCourtId(courtID3).
						WithElo(1250),
					models.NewDBRankingFixture().
						WithUserId(otherUser2).
						WithCourtId(courtID3).
						WithElo(1100),
				},
			},
			param: userID,
			expected: expected{
				fields: []models.Field{
					models.NewFieldFixture().
						WithRanking(1).
						WithName("Stade de Lyon").
						WithScore(1300),
					models.NewFieldFixture().
						WithRanking(1).
						WithName("Playground").
						WithScore(1250),
					models.NewFieldFixture().
						WithRanking(3).
						WithName("Central Park").
						WithScore(1200),
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
				return fields[i].Name < fields[j].Name
			})
			sort.Slice(c.expected.fields, func(i, j int) bool {
				return c.expected.fields[i].Name < c.expected.fields[j].Name
			})

			for i := range fields {
				require.Equal(t, c.expected.fields[i].Ranking, fields[i].Ranking, "wrong ranking for %s", fields[i].Name)
				require.Equal(t, c.expected.fields[i].Name, fields[i].Name)
				require.Equal(t, c.expected.fields[i].Score, fields[i].Score)
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

	initialRanking := models.NewDBRankingFixture().
		WithUserId(user.Id).
		WithCourtId(court.Id).
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
			name: "Update existing ranking on conflict",
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

			ranking, err := s.db.GetRankingByUserAndCourt(ctx, c.param.UserID, c.param.CourtID)
			require.NoError(t, err)
			require.NotNil(t, ranking)
			require.Equal(t, c.expected, ranking.Elo)
		})
	}
}

func TestDatabase_GetRankingByUserAndCourt(t *testing.T) {
	type expected struct {
		isError bool
		elo     int
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
				isError: false,
				elo:     1500,
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
				isError: true,
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
				isError: true,
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

			if c.expected.isError {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, c.userID, got.UserID)
			require.Equal(t, c.courtID, got.CourtID)
			require.Equal(t, c.expected.elo, got.Elo)
		})
	}
}
