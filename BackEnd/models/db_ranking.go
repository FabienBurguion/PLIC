package models

import "time"

type DBRanking struct {
	UserID    string    `db:"user_id"`
	CourtID   string    `db:"court_id"`
	Elo       int       `db:"elo"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
