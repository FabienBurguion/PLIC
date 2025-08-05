package models

import (
	"github.com/google/uuid"
	"time"
)

type MatchRequest struct {
	Sport           Sport     `json:"sport"`
	CourtID         string    `json:"court_id"`
	Date            time.Time `json:"date"`
	NbreParticipant int       `json:"nbre_participant"`
}

func (m MatchRequest) ToDBMatches(now time.Time) DBMatches {
	return DBMatches{
		Id:              uuid.NewString(),
		Sport:           m.Sport,
		Date:            m.Date,
		ParticipantNber: m.NbreParticipant,
		CurrentState:    ManqueJoueur,
		Score1:          0,
		Score2:          0,
		CourtID:         m.CourtID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

type MatchResponse struct {
	Id              string         `json:"id"`
	Sport           Sport          `json:"sport"`
	Place           string         `json:"place"`
	Date            time.Time      `json:"date"`
	NbreParticipant int            `json:"nbre_participant"`
	CurrentState    MatchState     `json:"current_state"`
	Score1          int            `json:"score1"`
	Score2          int            `json:"score2"`
	Users           []UserResponse `json:"users"`
	CreatedAt       time.Time      `json:"created_at"`
}
