package models

import "time"

type UserResponse struct {
	Username       string    `json:"username"`
	ProfilePicture *string   `json:"profilePicture"`
	Bio            *string   `json:"bio"`
	CreatedAt      time.Time `json:"createdAt"`
	VisitedFields  int       `json:"visitedFields"`
	Winrate        int       `json:"winrate"`
	FavoriteCity   string    `json:"favoriteCity"`
	FavoriteSport  Sport     `json:"favoriteSport"`
	FavoriteField  string    `json:"favoriteField"`
	Sports         []Sport   `json:"sports"`
	Fields         []Field   `json:"fields"`
}

type UserPatchRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Bio      *string `json:"bio"`
}
