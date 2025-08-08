package models

import (
	"time"

	"github.com/google/uuid"
)

type Sport string

const (
	Basket   Sport = "basket"
	Foot     Sport = "foot"
	PingPong Sport = "ping-pong"
)

type MatchState string

const (
	Termine      MatchState = "Termine"
	ManqueScore  MatchState = "Manque Score"
	EnCours      MatchState = "En cours"
	Valide       MatchState = "Valide"
	ManqueJoueur MatchState = "Manque joueur"
)

type DBMatches struct {
	Id              string     `db:"id"`
	Sport           Sport      `db:"sport"`
	Date            time.Time  `db:"date"`
	ParticipantNber int        `db:"participant_nber"`
	CurrentState    MatchState `db:"current_state"`
	Score1          *int       `db:"score1"`
	Score2          *int       `db:"score2"`
	CourtID         string     `db:"court_id"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

func NewDBMatchesFixture() DBMatches {
	return DBMatches{
		Id:              uuid.NewString(),
		Sport:           "foot",
		Date:            time.Now(),
		ParticipantNber: 8,
		CurrentState:    "Manque joueur",
		Score1:          nil,
		Score2:          nil,
	}
}

func (m DBMatches) WithId(id string) DBMatches {
	m.Id = id
	return m
}

func (m DBMatches) WithCourtId(courtId string) DBMatches {
	m.CourtID = courtId
	return m
}

func (m DBMatches) WithParticipantNber(participantNber int) DBMatches {
	m.ParticipantNber = participantNber
	return m
}

func (m DBMatches) WithCurrentState(currentState MatchState) DBMatches {
	m.CurrentState = currentState
	return m
}

func (m DBMatches) WithSport(sport Sport) DBMatches {
	m.Sport = sport
	return m
}

func (m DBMatches) WithScore1(score1 int) DBMatches {
	m.Score1 = &score1
	return m
}

func (m DBMatches) WithScore2(score2 int) DBMatches {
	m.Score2 = &score2
	return m
}
