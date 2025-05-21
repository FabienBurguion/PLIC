package models

type Field struct {
	Ranking int    `json:"ranking"`
	Name    string `json:"name"`
	Score   int    `json:"score"`
}
