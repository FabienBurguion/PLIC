package models

import (
	"github.com/google/uuid"
	"time"
)

type MatchRequest struct {
	Sport Sport     `json:"sport"`
	Place string    `json:"place"`
	Date  time.Time `json:"date"`
}

func (m MatchRequest) ToDBMatches() DBMatches {
	return DBMatches{
		Id:              uuid.NewString(),
		Sport:           m.Sport,
		Place:           m.Place,
		Date:            m.Date,
		ParticipantNber: 1,
		CurrentState:    ManqueJoueur,
		Score1:          0,
		Score2:          0,
	}
}

type MatchResponse struct {
	Id              string         `json:"id"`
	Sport           Sport          `json:"sport"`
	Place           string         `json:"place"`
	Date            time.Time      `json:"date"`
	ParticipantNber int            `json:"participant_nber"`
	CurrentState    EtatMatch      `json:"current_state"`
	Score1          int            `json:"score1"`
	Score2          int            `json:"score2"`
	Users           []UserResponse `json:"users"`
}
