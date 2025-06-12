package models

import (
	"github.com/google/uuid"
	"time"
)

type DBUsers struct {
	Id        string    `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Bio       *string   `db:"bio"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewDBUsersFixture() DBUsers {
	return DBUsers{
		Id:        uuid.NewString(),
		Username:  "username",
		Email:     "an email",
		Bio:       ptr("A bio"),
		Password:  "password",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u DBUsers) WithId(id string) DBUsers {
	u.Id = id
	return u
}

func (u DBUsers) WithUsername(username string) DBUsers {
	u.Username = username
	return u
}

func (u DBUsers) WithPassword(password string) DBUsers {
	u.Password = password
	return u
}

func (u DBUsers) WithEmail(email string) DBUsers {
	u.Email = email
	return u
}

func (u DBUsers) WithBio(bio string) DBUsers {
	u.Bio = ptr(bio)
	return u
}

func (u DBUsers) WitUpdatedAt(updatedAt time.Time) DBUsers {
	u.UpdatedAt = updatedAt
	return u
}

func (u DBUsers) ToUserResponse(profilePictureUrl string) UserResponse {
	var p *string
	if profilePictureUrl != "" {
		p = &profilePictureUrl
	}
	return UserResponse{
		Username:       u.Username,
		Bio:            u.Bio,
		CreatedAt:      u.CreatedAt,
		ProfilePicture: p,   // Tu peux ajouter ton s3 logic ici si nécessaire
		VisitedFields:  0,   // TODO
		Winrate:        nil, // TODO
		FavoriteCity:   nil, // TODO
		FavoriteSport:  nil, // TODO
		FavoriteField:  nil, // TODO
		Sports:         nil, // TODO
		Fields:         nil, // TODO
	}
}
