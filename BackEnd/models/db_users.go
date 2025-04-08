package models

type DBUsers struct {
	Id        string `db:"id"`
	Username  string `db:"username"`
	hPassword string `db:"password"`
}
