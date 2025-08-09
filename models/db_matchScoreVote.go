package models

import "time"

type DBMatchScoreVote struct {
	MatchId   string    `db:"match_id"`
	UserId    string    `db:"user_id"`
	Team      int       `db:"team"`
	Score1    int       `db:"score1"`
	Score2    int       `db:"score2"`
	CreatedAt time.Time `db:"created_at"`
}
