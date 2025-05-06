package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (db Database) CheckMatchExist(ctx context.Context, id string) (bool, error) {
	var match models.DBMatches

	err := db.Database.GetContext(ctx, &match, `
		SELECT id, sport, lieu, date, nbre_participant, etat, score1, score2
		FROM matches
		WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return true, nil
}

func (db Database) GetMatchById(ctx context.Context, id string) (*models.DBMatches, error) {
	var match models.DBMatches

	err := db.Database.GetContext(ctx, &match, `
        SELECT id, sport, lieu, date, nbre_participant, etat, score1, score2
        FROM matches
        WHERE id = $1`, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return &match, nil
}

func (db Database) GetAllMatches(ctx context.Context) ([]models.DBMatches, error) {
	var matches []models.DBMatches
	err := db.Database.SelectContext(ctx, &matches, `
        SELECT id, sport, lieu, date, nbre_participant, etat, score1, score2
        FROM matches`)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des matchs : %w", err)
	}
	return matches, nil
}

func (db Database) CreateMatch(ctx context.Context, match models.DBMatches) error {
	_, err := db.Database.NamedExecContext(ctx, `
        INSERT INTO matches (id, sport, lieu, date, nbre_participant, etat, score1, score2)
        VALUES (:id, :sport, :lieu, :date, :nbre_participant, :etat, :score1, :score2)`, match)
	if err != nil {
		return fmt.Errorf("échec de l'insertion match : %w", err)
	}
	return nil
}
