package models

import "time"

type GetMatchByCourtIdResponses struct {
	Id              string    `json:"id"`
	Sport           Sport     `json:"sport"`
	Place           string    `json:"place"`
	Date            time.Time `json:"date"`
	ParticipantNber int       `json:"participant_nber"`
	CurrentState    EtatMatch `json:"current_state"`
	Score1          int       `json:"score1"`
	Score2          int       `json:"score2"`
}
