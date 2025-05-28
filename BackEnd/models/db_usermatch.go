package models

import "time"

type DBUserMatch struct {
	UserID    string    `db:"user_id"`
	MatchID   string    `db:"match_id"`
	CreatedAt time.Time `db:"created_at"`
}
