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
