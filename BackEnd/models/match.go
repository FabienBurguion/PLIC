package models

import (
	"github.com/google/uuid"
	"time"
)

type MatchRequest struct {
	Sport Sport     `json:"sport"`
	Lieu  string    `json:"lieu"`
	Date  time.Time `json:"date"`
}

func (m MatchRequest) ToDBMatches() DBMatches {
	return DBMatches{
		Id:              uuid.NewString(),
		Sport:           m.Sport,
		Lieu:            m.Lieu,
		Date:            m.Date,
		NbreParticipant: 1,
		Etat:            ManqueJoueur,
		Score1:          0,
		Score2:          0,
	}
}

type MatchResponse struct {
	Id              string    `json:"id"`
	Sport           Sport     `json:"sport"`
	Lieu            string    `json:"lieu"`
	Date            time.Time `json:"date"`
	NbreParticipant int       `json:"nbre_participant"`
	Etat            EtatMatch `json:"etat"`
	Score1          int       `json:"score1"`
	Score2          int       `json:"score2"`
}
