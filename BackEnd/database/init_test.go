package database

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"io/ioutil"
	"os"
)

type Service struct {
	db Database
}

func (s *Service) InitServiceTest() {
	db, err := InitDBTest("../database/sql/1.0.0.sql")
	if err != nil {
		panic(err)
	}
	s.db = Database{
		Database: db,
	}
}

func InitDBTest(sqlFile string) (*sqlx.DB, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		panic(err)
	}
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		panic("DATABASE_URL environment variable is not set")
	}
	dsn := "host=" + dockerHost + " port=5433 user=test password=test dbname=test sslmode=disable"
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
