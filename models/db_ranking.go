package models

import (
	"time"

	"github.com/google/uuid"
)

type DBRanking struct {
	UserID    string    `db:"user_id"`
	CourtID   string    `db:"court_id"`
	Elo       int       `db:"elo"`
	Sport     Sport     `db:"sport"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewDBRankingFixture() DBRanking {
	return DBRanking{
		UserID:    uuid.NewString(),
		CourtID:   uuid.NewString(),
		Elo:       1000,
		Sport:     Basket,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u DBRanking) WithUserId(userId string) DBRanking {
	u.UserID = userId
	return u
}

func (u DBRanking) WithCourtId(courtId string) DBRanking {
	u.CourtID = courtId
	return u
}

func (u DBRanking) WithElo(elo int) DBRanking {
	u.Elo = elo
	return u
}

func (u DBRanking) WithSport(sport Sport) DBRanking {
	u.Sport = sport
	return u
}
