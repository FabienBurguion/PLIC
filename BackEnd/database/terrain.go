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

func (db Database) GetAllTerrains(ctx context.Context) ([]models.DBCourt, error) {
	var terrains []models.DBCourt
	err := db.Database.SelectContext(ctx, &terrains, `
		SELECT id, address, longitude, latitude, created_at, name
		FROM courts`)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des terrains : %w", err)
	}
	return terrains, nil
}
