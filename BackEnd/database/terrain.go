package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (db Database) InsertTerrain(ctx context.Context, id string, p models.Place) error {
	_, err := db.Database.ExecContext(ctx, `
		INSERT INTO terrain (id, address, longitude, latitude)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`,
		id, p.Address, p.Geometry.Location.Lat, p.Geometry.Location.Lng,
	)
	if err != nil {
		return fmt.Errorf("échec de l'insertion du terrain : %w", err)
	}

	return nil
}

func (db Database) GetTerrainByAddress(ctx context.Context, address string) (*models.DBCourt, error) {
	var court models.DBCourt

	err := db.Database.GetContext(ctx, &court, `
		SELECT id, address, longitude, latitude
		FROM terrain
		WHERE address = $1`, address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return &court, nil
}
