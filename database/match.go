package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

func (db Database) CheckMatchExist(ctx context.Context, id string) (bool, error) {
	var match models.DBMatches

	err := db.Database.GetContext(ctx, &match, `
		SELECT id, sport, place, date, participant_nber, current_state, score1, score2, created_at, updated_at
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
        SELECT id, sport, place, date, participant_nber, current_state, score1, score2, court_id, created_at, updated_at
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
        SELECT u.id, u.username, u.email, u.bio, u.current_field_id, u.password, u.created_at, u.updated_at
        FROM user_match um
        JOIN users u ON um.user_id = u.id
        WHERE um.match_id = $1`, matchId)

	if err != nil {
		fmt.Println("error is here")
		return nil, fmt.Errorf("échec de la récupération des utilisateurs du match %s : %w", matchId, err)
	}

	return users, nil
}

func (db Database) GetMatchesByUserID(ctx context.Context, userID string) ([]models.DBMatchByUserId, error) {
	var dbMatches []models.DBMatchByUserId
	err := db.Database.SelectContext(ctx, &dbMatches, `
		SELECT m.id, m.sport, m.place, m.date, m.participant_nber, m.current_state, m.score1, m.score2
		FROM matches m
		JOIN user_match um ON m.id = um.match_id
		WHERE um.user_id = $1
		ORDER BY m.date DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying matches for user: %w", err)
	}
	return dbMatches, nil
}

func (db Database) GetMatchesByCourtId(ctx context.Context, courtID string) ([]models.DBMatches, error) {
	var dbMatches []models.DBMatches
	log.Printf("🟡 Requête pour court %s à %v", courtID, time.Now())
	err := db.Database.SelectContext(ctx, &dbMatches, `
        SELECT id, sport, place, date, participant_nber, current_state, score1, score2, court_id, created_at, updated_at
        FROM matches
        WHERE court_id = $1
        ORDER BY date DESC
    `, courtID)
	if err != nil {
		msg := fmt.Errorf("error querying matches for court %s: %w", courtID, err)
		log.Println(msg)
		return nil, msg
	}

	return dbMatches, nil
}

func (db Database) GetMatchCountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	err := db.Database.GetContext(ctx, &count, `
		SELECT COUNT(*) AS match_count
		FROM user_match um
		JOIN matches m ON um.match_id = m.id
		WHERE um.user_id = $1
		AND m.current_state IN ('Termine', 'Manque Score')
	`, userID)
	if err != nil {
		return 0, fmt.Errorf("error counting finished matches for user: %w", err)
	}
	return count, nil
}

func (db Database) GetAllMatches(ctx context.Context) ([]models.DBMatches, error) {
	var matches []models.DBMatches
	err := db.Database.SelectContext(ctx, &matches, `
        SELECT id, sport, place, date, participant_nber, current_state, score1, score2, created_at, updated_at
        FROM matches`)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des matchs : %w", err)
	}
	return matches, nil
}

func (db Database) CreateMatch(ctx context.Context, match models.DBMatches) error {
	_, err := db.Database.NamedExecContext(ctx, `
    INSERT INTO matches (
        id, sport, place, date, participant_nber, current_state, score1, score2, court_id, created_at, updated_at
    ) VALUES (
        :id, :sport, :place, :date, :participant_nber, :current_state, :score1, :score2, :court_id, :created_at, :updated_at
    )`, match)

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

func (db Database) DeleteMatch(ctx context.Context, matchID string) error {
	_, err := db.Database.ExecContext(ctx, `
		DELETE FROM matches
		WHERE id = $1
	`, matchID)

	if err != nil {
		return fmt.Errorf("échec de la suppression du match %s : %w", matchID, err)
	}
	return nil
}

func (db Database) CreateUserMatch(ctx context.Context, um models.DBUserMatch) error {
	_, err := db.Database.ExecContext(ctx,
		`INSERT INTO user_match (user_id, match_id, created_at) VALUES ($1, $2, $3)`,
		um.UserID, um.MatchID, um.CreatedAt)
	return err
}

func (db Database) UpdateMatchScore(ctx context.Context, id string, score1, score2 int, updatedTime time.Time) error {
	_, err := db.Database.ExecContext(ctx, `
		UPDATE matches 
		SET score1 = $1, score2 = $2, current_state = $3, updated_at = $4
		WHERE id = $5
	`, score1, score2, models.Termine, updatedTime, id)

	if err != nil {
		return fmt.Errorf("échec mise à jour du score: %w", err)
	}

	return nil
}
