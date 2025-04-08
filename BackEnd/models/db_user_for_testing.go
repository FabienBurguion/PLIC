package models

type DBUserTest struct {
	Id    string `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}
