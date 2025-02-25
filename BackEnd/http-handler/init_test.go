package main

import (
	"PLIC/database"
	"database/sql"
	_ "github.com/lib/pq"
	"io/ioutil"
)

func (s *Service) InitServiceTest() {
	db, err := InitDBTest("database/sql/1.0.0.sql")
	if err != nil {
		panic(err)
	}
	s.db = database.Database{
		Database: db,
	}
}

func InitDBTest(sqlFile string) (*sql.DB, error) {
	dsn := "host=localhost port=5433 user=test password=test dbname=test sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	sqlBytes, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return nil, err
	}

	return db, nil
}
