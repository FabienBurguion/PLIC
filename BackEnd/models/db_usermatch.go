package models

import "time"

type DBUserMatch struct {
	UserID    string    `db:"user_id"`
	MatchID   string    `db:"match_id"`
	CreatedAt time.Time `db:"created_at"`
}

type DBMatchByUserId struct {
	Id              string    `db:"id"`
	Sport           Sport     `db:"sport"`
	Place           string    `db:"place"`
	Date            time.Time `db:"date"`
	ParticipantNber int       `db:"participant_nber"`
	CurrentState    EtatMatch `db:"current_state"`
	Score1          int       `db:"score1"`
	Score2          int       `db:"score2"`
}
