package models

import "time"

type Sport string

const (
	Basket Sport = "basket"
	Foot   Sport = "foot"
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
}

func (m DBMatches) ToMatchResponse(users []DBUsers) MatchResponse {
	userResponses := make([]UserResponse, len(users))
	for i, u := range users {
		userResponses[i] = u.ToUserResponse() // crée cette méthode si pas encore existante
	}

	return MatchResponse{
		Id:              m.Id,
		Sport:           m.Sport,
		Place:           m.Place,
		Date:            m.Date,
		ParticipantNber: m.ParticipantNber,
		CurrentState:    m.CurrentState,
		Score1:          m.Score1,
		Score2:          m.Score2,
		Users:           userResponses,
	}
}
