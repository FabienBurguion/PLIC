package models

import (
	"time"

	"github.com/google/uuid"
)

type DBUserMatch struct {
	UserID    string    `db:"user_id"`
	MatchID   string    `db:"match_id"`
	Team      int       `db:"team"`
	CreatedAt time.Time `db:"created_at"`
}

func NewDBUserMatchFixture() DBUserMatch {
	return DBUserMatch{
		UserID:    uuid.NewString(),
		MatchID:   uuid.NewString(),
		Team:      1,
		CreatedAt: time.Now(),
	}
}

func (u DBUserMatch) WithUserId(userId string) DBUserMatch {
	u.UserID = userId
	return u
}

func (u DBUserMatch) WithMatchId(matchId string) DBUserMatch {
	u.MatchID = matchId
	return u
}

func (u DBUserMatch) WithTeam(team int) DBUserMatch {
	u.Team = team
	return u
}
