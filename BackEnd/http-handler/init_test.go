package main

import (
	"PLIC/database"
	"PLIC/models"
	"PLIC/s3_management"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DBFixtures struct {
	Users []models.DBUsers
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
	db, cleanup, err := InitDBTest(file)
	if err != nil {
		panic(err)
	}
	s.db = database.Database{
		Database: db,
	}

	go func() {
		<-time.After(10 * time.Second)
		_ = cleanup()
	}()

	mockS3 := &s3_management.MockS3Service{}

	s.s3Service = mockS3
}

func InitDBTest(sqlFile string) (*sqlx.DB, func() error, error) {
	dockerHost := "localhost"
	/*
		if dockerHost == "" {
			panic("DOCKER_HOST environment variable is not set")
		}
	*/
	adminDsn := "host=" + dockerHost + " port=5433 user=test password=test dbname=postgres sslmode=disable"

	adminDb, err := sql.Open("postgres", adminDsn)
	if err != nil {
		return nil, nil, err
	}

	uuidStr := strings.ReplaceAll(uuid.NewString(), "-", "")
	dbName := "test_" + uuidStr

	_, err = adminDb.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		return nil, nil, err
	}

	testDsn := fmt.Sprintf("host=%s port=5433 user=test password=test dbname=%s sslmode=disable", dockerHost, dbName)
	testDb, err := sql.Open("postgres", testDsn)
	if err != nil {
		return nil, nil, err
	}

	sqlBytes, err := os.ReadFile(sqlFile)
	if err != nil {
		return nil, nil, err
	}

	_, err = testDb.Exec(string(sqlBytes))
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		_ = testDb.Close()
		_, err := adminDb.Exec("DROP DATABASE IF EXISTS " + dbName)
		_ = adminDb.Close()
		return err
	}

	return sqlx.NewDb(testDb, "postgres"), cleanup, nil
}

func (s *Service) loadFixtures(fixtures DBFixtures) {
	ctx := context.Background()
	for _, u := range fixtures.Users {
		err := s.db.CreateUser(ctx, u)
		if err != nil {
			panic(err)
		}
	}
}
