package main

import (
	"PLIC/database"
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"io/ioutil"
)

func (s *Service) InitServiceTest() {
	db, err := InitDBTest("../database/sql/1.0.0.sql")
	if err != nil {
		panic(err)
	}
	s.db = database.Database{
		Database: db,
	}
}

func InitDBTest(sqlFile string) (*sqlx.DB, error) {
	dsn := "host=host.docker.internal port=5433 user=test password=test dbname=test sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	sqlBytes, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		DO $$ 
		DECLARE 
			r RECORD;
		BEGIN 
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') 
			LOOP 
				EXECUTE 'DROP TABLE IF EXISTS public.' || r.tablename || ' CASCADE';
			END LOOP; 
		END $$;
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return nil, err
	}

	return sqlx.NewDb(db, "postgres"), nil
}
