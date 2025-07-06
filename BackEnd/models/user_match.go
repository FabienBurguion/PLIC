package models

import "time"

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

type GetMatchByUserIdResponses struct {
	Id              string    `json:"id"`
	Sport           Sport     `json:"sport"`
	Place           string    `json:"place"`
	Date            time.Time `json:"date"`
	ParticipantNber int       `json:"participant_nber"`
	CurrentState    EtatMatch `json:"current_state"`
	Score1          int       `json:"score1"`
	Score2          int       `json:"score2"`
}
