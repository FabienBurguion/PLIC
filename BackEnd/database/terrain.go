package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (db Database) GetPlaces(latitude, longitude float64, apiKey string) ([]models.Place, error) {
	url := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/place/nearbysearch/json?location=%f,%f&radius=1000&type=sports_complex&key=%s",
		latitude, longitude, apiKey)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data models.GooglePlacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Results, nil
}

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
