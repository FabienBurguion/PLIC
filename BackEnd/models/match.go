package models

import "time"

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
