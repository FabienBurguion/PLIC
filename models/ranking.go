package models

type CourtRankingResponse struct {
	UserID string `json:"userId" db:"user_id"`
	Elo    int    `json:"elo"    db:"elo"`
}
