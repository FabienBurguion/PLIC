package models

type UpdateScoreRequest struct {
	Score1 int `json:"score1"`
	Score2 int `json:"score2"`
}

func NewUpdateScoreRequestFixture() UpdateScoreRequest {
	return UpdateScoreRequest{
		Score1: 0,
		Score2: 0,
	}
}

func (u UpdateScoreRequest) WithScore1(score1 int) UpdateScoreRequest {
	u.Score1 = score1
	return u
}

func (u UpdateScoreRequest) WithScore2(score2 int) UpdateScoreRequest {
	u.Score2 = score2
	return u
}
