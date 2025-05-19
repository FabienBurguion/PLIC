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
	Lieu            string    `db:"lieu"`
	Date            time.Time `db:"date"`
	NbreParticipant int       `db:"nbre_participant" json:"nbre_participant"`
	Etat            EtatMatch `db:"etat"`
	Score1          int       `db:"score1" json:"score1"`
	Score2          int       `db:"score2" json:"score2"`
}

func (m DBMatches) ToMatchResponse() MatchResponse {
	return MatchResponse{
		Id:              m.Id,
		Sport:           m.Sport,
		Lieu:            m.Lieu,
		Date:            m.Date,
		NbreParticipant: m.NbreParticipant,
		Etat:            m.Etat,
		Score1:          m.Score1,
		Score2:          m.Score2,
	}
}
