package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (db Database) InsertTerrain(ctx context.Context, id string, p models.Place, createdTime time.Time) error {
	_, err := db.Database.ExecContext(ctx, `
		INSERT INTO courts (id, address, longitude, latitude, created_at, name)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`,
		id, p.Address, p.Geometry.Location.Lng, p.Geometry.Location.Lat, createdTime, p.Name,
	)
	if err != nil {
		return fmt.Errorf("échec de l'insertion du terrain : %w", err)
	}

	return nil
}

func (db Database) GetTerrainByAddress(ctx context.Context, address string) (*models.DBCourt, error) {
	var court models.DBCourt

	err := db.Database.GetContext(ctx, &court, `
		SELECT id, address, longitude, latitude, created_at, name
		FROM courts
		WHERE address = $1`, address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return &court, nil
}

func (db Database) GetAllCourts(ctx context.Context) ([]models.DBCourt, error) {
	var terrains []models.DBCourt
	err := db.Database.SelectContext(ctx, &terrains, `
		SELECT id, address, longitude, latitude, created_at, name
		FROM courts`)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des terrains : %w", err)
	}
	return terrains, nil
}

func (db Database) GetVisitedFieldCountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	err := db.Database.GetContext(ctx, &count, `
		SELECT COUNT(DISTINCT m.place)
		FROM user_match um
		JOIN matches m ON m.id = um.match_id
		WHERE um.user_id = $1
	`, userID)
	if err != nil {
		return 0, fmt.Errorf("error counting visited fields: %w", err)
	}
	return count, nil
}

func (db Database) GetCourtByID(ctx context.Context, id string) (*models.DBCourt, error) {
	var court models.DBCourt
	err := db.Database.GetContext(ctx, &court, `
		SELECT id, address, name, longitude, latitude, created_at
		FROM courts
		WHERE id = $1
	`, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch court: %w", err)
	}
	return &court, nil
}

func (db Database) InsertCourtForTest(ctx context.Context, court models.DBCourt) error {
	_, err := db.Database.NamedExecContext(ctx, `
		INSERT INTO courts (id, name, address, latitude, longitude, created_at)
		VALUES (:id, :name, :address, :latitude, :longitude, :created_at)`, court)
	return err
}

func (db Database) CreateCourt(ctx context.Context, court models.DBCourt) error {
	_, err := db.Database.ExecContext(ctx, `
		INSERT INTO courts (id, name, address, longitude, latitude, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, court.Id, court.Name, court.Address, court.Longitude, court.Latitude, court.CreatedAt)

	if err != nil {
		return fmt.Errorf("échec de l'insertion court : %w", err)
	}
	return nil
}
