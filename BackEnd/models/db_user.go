package models

type DBUser struct {
	Id    string `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}
