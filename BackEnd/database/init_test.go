package database

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

type Service struct {
	db Database
}

func findLatestMigrationFile(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.sql$`)

	type versionedFile struct {
		name    string
		version [3]int
	}

	var candidates []versionedFile

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(file.Name())
		if matches == nil {
			continue
		}

		major, _ := strconv.Atoi(matches[1])
		minor, _ := strconv.Atoi(matches[2])
		patch, _ := strconv.Atoi(matches[3])

		candidates = append(candidates, versionedFile{
			name:    file.Name(),
			version: [3]int{major, minor, patch},
		})
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no valid migration files found in %s", dir)
	}

	sort.Slice(candidates, func(i, j int) bool {
		for k := 0; k < 3; k++ {
			if candidates[i].version[k] != candidates[j].version[k] {
				return candidates[i].version[k] < candidates[j].version[k]
			}
		}
		return false
	})

	latest := candidates[len(candidates)-1].name
	return filepath.Join(dir, latest), nil
}

func (s *Service) InitServiceTest() {
	file, err := findLatestMigrationFile("../database/sql")
	if err != nil {
		panic(err)
	}
	db, err := InitDBTest(file)
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
		panic("DOCKER_HOST environment variable is not set")
	}
	dsn := "host=" + dockerHost + " port=5433 user=test password=test dbname=test sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	sqlBytes, err := os.ReadFile(sqlFile)
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
