package models

type Field struct {
	Ranking int    `json:"ranking"`
	Name    string `json:"name"`
	Score   int    `json:"score"`
}

func NewFieldFixture() Field {
	return Field{
		Ranking: 0,
		Name:    "",
		Score:   0,
	}
}

func (f Field) WithRanking(ranking int) Field {
	f.Ranking = ranking
	return f
}

func (f Field) WithName(name string) Field {
	f.Name = name
	return f
}

func (f Field) WithScore(score int) Field {
	f.Score = score
	return f
}
