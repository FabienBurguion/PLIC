package models

import "time"

type UserResponse struct {
	Username string `json:"username"`
	// @nullable
	ProfilePicture *string `json:"profilePicture"`
	// @nullable
	Bio            *string   `json:"bio"`
	CurrentFieldId *string   `json:"currentFieldId"`
	CreatedAt      time.Time `json:"createdAt"`
	VisitedFields  int       `json:"visitedFields"`
	NbMatches      int       `json:"nbMatches"`
	// @nullable
	Winrate *int `json:"winrate"`
	// @nullable
	FavoriteCity *string `json:"favoriteCity"`
	// @nullable
	FavoriteSport *Sport `json:"favoriteSport"`
	// @nullable
	FavoriteField *string `json:"favoriteField"`
	Sports        []Sport `json:"sports"`
	Fields        []Field `json:"fields"`
}

type UserPatchRequest struct {
	// @nullable
	Username *string `json:"username"`
	// @nullable
	Email *string `json:"email"`
	// @nullable
	Bio *string `json:"bio"`
	// @nullable
	CurrentFieldId *string `json:"currentFieldId"`
}

type UserStats struct {
	MatchCount    int
	VisitedFields int
	FavoriteSport *Sport
	FavoriteField *string
	Sports        []Sport
	Fields        []Field
	Winrate       *int
}
