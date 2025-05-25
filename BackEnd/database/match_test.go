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
	ctx := context.Background()
	s := &Service{}
	s.InitServiceTest()

	id := uuid.NewString()
	match := models.DBMatches{
		Id:              id,
		Sport:           "basket",
		Lieu:            "Paris",
		Date:            time.Now(),
		NbreParticipant: 8,
		Etat:            "Manque joueur",
		Score1:          0,
		Score2:          0,
	}

	err := s.db.CreateMatch(ctx, match)
	require.NoError(t, err)

	dbMatch, err := s.db.GetMatchById(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, dbMatch)
	require.Equal(t, match.Id, dbMatch.Id)
	require.Equal(t, match.Lieu, dbMatch.Lieu)
}

func TestDatabase_GetMatchById(t *testing.T) {
	ctx := context.Background()
	s := &Service{}
	s.InitServiceTest()

	id := uuid.NewString()
	match := models.DBMatches{
		Id:              id,
		Sport:           "foot",
		Lieu:            "Lyon",
		Date:            time.Now(),
		NbreParticipant: 10,
		Etat:            "Manque joueur",
		Score1:          1,
		Score2:          2,
	}

	// TEST TO DELETE

	err := s.db.CreateMatch(ctx, match)
	require.NoError(t, err)

	result, err := s.db.GetMatchById(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, match.Id, result.Id)
	require.Equal(t, match.Score2, result.Score2)
}

func TestDatabase_GetAllMatches(t *testing.T) {
	ctx := context.Background()
	s := &Service{}
	s.InitServiceTest()

	id1 := uuid.NewString()
	id2 := uuid.NewString()

	match1 := models.DBMatches{Id: id1, Sport: "foot", Lieu: "Nice", Date: time.Now(), NbreParticipant: 2, Etat: "Manque joueur", Score1: 0, Score2: 0}
	match2 := models.DBMatches{Id: id2, Sport: "basket", Lieu: "Paris", Date: time.Now(), NbreParticipant: 10, Etat: "Manque joueur", Score1: 0, Score2: 0}

	err := s.db.CreateMatch(ctx, match1)
	require.NoError(t, err)
	err = s.db.CreateMatch(ctx, match2)
	require.NoError(t, err)

	matches, err := s.db.GetAllMatches(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(matches), 2)

	var found1, found2 bool
	for _, m := range matches {
		if m.Id == id1 {
			found1 = true
		}
		if m.Id == id2 {
			found2 = true
		}
	}
	require.True(t, found1)
	require.True(t, found2)
}
