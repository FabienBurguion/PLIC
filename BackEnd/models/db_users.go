package models

type DBUsers struct {
	Id       string `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}
