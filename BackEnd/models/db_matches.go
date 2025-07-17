package models

import (
	"github.com/google/uuid"
	"time"
)

type Sport string

const (
	Basket   Sport = "basket"
	Foot     Sport = "foot"
	PingPong Sport = "ping-pong"
)

type EtatMatch string

const (
	Termine      EtatMatch = "Termine"
	ManqueScore  EtatMatch = "Manque Score"
	EnCours      EtatMatch = "En cours"
	Valide       EtatMatch = "Valide"
	ManqueJoueur EtatMatch = "Manque joueur"
)

type DBMatches struct {
	Id              string    `db:"id"`
	Sport           Sport     `db:"sport"`
	Place           string    `db:"place"`
	Date            time.Time `db:"date"`
	ParticipantNber int       `db:"participant_nber"`
	CurrentState    EtatMatch `db:"current_state"`
	Score1          int       `db:"score1"`
	Score2          int       `db:"score2"`
	CourtID         string    `db:"court_id"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

func NewDBMatchesFixture() DBMatches {
	return DBMatches{
		Id:              uuid.NewString(),
		Sport:           "foot",
		Place:           "Paris",
		Date:            time.Now(),
		ParticipantNber: 8,
		CurrentState:    "Manque joueur",
		Score1:          0,
		Score2:          0,
	}
}
