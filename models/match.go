package models

import (
	"time"

	"github.com/google/uuid"
)

type MatchRequest struct {
	Sport           Sport     `json:"sport"`
	CourtID         string    `json:"court_id"`
	Date            time.Time `json:"date"`
	NbreParticipant int       `json:"nbre_participant"`
}

func NewMatchRequestFixture() MatchRequest {
	return MatchRequest{
		Sport:           Foot,
		CourtID:         uuid.NewString(),
		Date:            time.Now(),
		NbreParticipant: 2,
	}
}

func (m MatchRequest) WithCourtId(courtId string) MatchRequest {
	m.CourtID = courtId
	return m
}

func (m MatchRequest) ToDBMatches(now time.Time) DBMatches {
	return DBMatches{
		Id:              uuid.NewString(),
		Sport:           m.Sport,
		Date:            m.Date,
		ParticipantNber: m.NbreParticipant,
		CurrentState:    ManqueJoueur,
		Score1:          nil,
		Score2:          nil,
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
	Score1          *int           `json:"score1"`
	Score2          *int           `json:"score2"`
	Users           []UserResponse `json:"users"`
	CreatedAt       time.Time      `json:"created_at"`
}

type JoinMatchRequest struct {
	Team int `json:"team"`
}

type CreateMatchResponse struct {
	Id string `json:"id"`
}
