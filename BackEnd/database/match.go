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
		SELECT id, sport, place, date, participant_nber, current_state, score1, score2
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
        SELECT id, sport, place, date, participant_nber, current_state, score1, score2
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

func (db Database) GetUsersByMatchId(ctx context.Context, matchId string) ([]models.DBUsers, error) {
	var users []models.DBUsers
	err := db.Database.SelectContext(ctx, &users, `
        SELECT u.id, u.username, u.email, u.bio, u.password, u.created_at, u.updated_at
        FROM user_match um
        JOIN users u ON um.user_id = u.id
        WHERE um.match_id = $1`, matchId)

	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des utilisateurs du match %s : %w", matchId, err)
	}

	return users, nil
}

func (db Database) GetAllMatches(ctx context.Context) ([]models.DBMatches, error) {
	var matches []models.DBMatches
	err := db.Database.SelectContext(ctx, &matches, `
        SELECT id, sport, place, date, participant_nber, current_state, score1, score2
        FROM matches`)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des matchs : %w", err)
	}
	return matches, nil
}

func (db Database) CreateMatch(ctx context.Context, match models.DBMatches) error {
	_, err := db.Database.NamedExecContext(ctx, `
        INSERT INTO matches (id, sport, place, date, participant_nber, current_state, score1, score2)
        VALUES (:id, :sport, :place, :date, :participant_nber, :current_state, :score1, :score2)`, match)
	if err != nil {
		return fmt.Errorf("échec de l'insertion match : %w", err)
	}
	return nil
}

func (db Database) AddUserToMatch(ctx context.Context, um models.DBUserMatch) error {
	_, err := db.Database.NamedExecContext(ctx, `
        INSERT INTO user_match (user_id, match_id, created_at)
        VALUES (:user_id, :match_id, :created_at)
    `, um)
	if err != nil {
		return fmt.Errorf("échec de l'ajout du user au match : %w", err)
	}
	return nil
}

func (db Database) IsUserInMatch(ctx context.Context, userID, matchID string) (bool, error) {
	var exists struct{}
	err := db.Database.GetContext(ctx, &exists, `
        SELECT 1
        FROM user_match
        WHERE user_id = $1 AND match_id = $2
        LIMIT 1
    `, userID, matchID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("erreur lors de la vérification de user_match: %w", err)
	}
	return true, nil
}

func (db Database) IncrementMatchParticipants(ctx context.Context, matchID string) error {
	_, err := db.Database.ExecContext(ctx, `
        UPDATE matches
        SET participant_nber = participant_nber + 1
        WHERE id = $1
    `, matchID)
	if err != nil {
		return fmt.Errorf("erreur lors de l'incrémentation du nombre de participants: %w", err)
	}
	return nil
}
