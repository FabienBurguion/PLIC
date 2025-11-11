package models

type Field struct {
	Ranking int    `json:"ranking"`
	Name    string `json:"name"`
	Elo     int    `json:"elo"`
	Sport   Sport  `json:"sport"`
}

func NewFieldFixture() Field {
	return Field{
		Ranking: 0,
		Name:    "",
		Elo:     0,
		Sport:   Basket,
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
	f.Elo = score
	return f
}

func (f Field) WithSport(sport Sport) Field {
	f.Sport = sport
	return f
}
