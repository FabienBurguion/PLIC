package database

import "github.com/jmoiron/sqlx"

type Database struct {
	Database *sqlx.DB
}
