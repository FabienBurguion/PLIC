package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (db Database) GetMatchById(ctx context.Context, id string) (*models.DBMatches, error) {
	var match models.DBMatches

	err := db.Database.GetContext(ctx, &match, `
        SELECT id, sport, date, participant_nber, current_state, score1, score2, creator_id, court_id, created_at, updated_at
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
		return nil, fmt.Errorf("échec de la récupération des utilisateurs du param %s : %w", matchId, err)
	}
	return users, nil
}

func (db Database) GetMatchesByUserID(ctx context.Context, userID string) ([]models.DBMatches, error) {
	var dbMatches []models.DBMatches
	err := db.Database.SelectContext(ctx, &dbMatches, `
		SELECT m.id, m.sport, m.date, m.participant_nber, m.current_state, m.score1, m.score2, m.court_id, m.creator_id, m.created_at, m.updated_at
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
	err := db.Database.SelectContext(ctx, &dbMatches, `
        SELECT id, sport, date, participant_nber, current_state, score1, score2, court_id, creator_id, created_at, updated_at
        FROM matches
        WHERE court_id = $1
        ORDER BY date DESC
    `, courtID)
	if err != nil {
		msg := fmt.Errorf("error querying matches for court %s: %w", courtID, err)
		return nil, msg
	}

	return dbMatches, nil
}

func (db Database) GetMatchCountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	err := db.Database.GetContext(ctx, &count, `
		SELECT COALESCE(COUNT(*), 0) AS match_count
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
        SELECT id, sport, date, participant_nber, current_state, score1, score2, court_id, creator_id, created_at, updated_at
        FROM matches`)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des matchs : %w", err)
	}
	return matches, nil
}

func (db Database) CreateMatch(ctx context.Context, match models.DBMatches) error {
	_, err := db.Database.NamedExecContext(ctx, `
    INSERT INTO matches (
        id, sport, date, participant_nber, current_state, score1, score2, court_id, creator_id, created_at, updated_at
    ) VALUES (
        :id, :sport, :date, :participant_nber, :current_state, :score1, :score2, :court_id, :creator_id, :created_at, :updated_at
    )`, match)

	if err != nil {
		return fmt.Errorf("échec de l'insertion : %w", err)
	}
	return nil
}

func (db Database) IsUserInMatch(ctx context.Context, userID, matchID string) (bool, error) {
	var dummy int
	err := db.Database.GetContext(ctx, &dummy, `
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
		return fmt.Errorf("échec de la suppression du param %s : %w", matchID, err)
	}
	return nil
}

func (db Database) CreateUserMatch(ctx context.Context, um models.DBUserMatch) error {
	_, err := db.Database.ExecContext(ctx,
		`INSERT INTO user_match (user_id, match_id, team, created_at) VALUES ($1, $2, $3, $4)`,
		um.UserID, um.MatchID, um.Team, um.CreatedAt)
	return err
}

func (db Database) UpsertMatch(ctx context.Context, match models.DBMatches, now time.Time) error {
	_, err := db.Database.ExecContext(ctx, `
		INSERT INTO matches (id, sport, date, participant_nber, current_state, score1, score2, court_id, creator_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			sport = EXCLUDED.sport,
			date = EXCLUDED.date,
			participant_nber = EXCLUDED.participant_nber,
			current_state = EXCLUDED.current_state,
			score1 = EXCLUDED.score1,
			score2 = EXCLUDED.score2,
			court_id = EXCLUDED.court_id,
			creator_id = EXCLUDED.creator_id,
			updated_at = $11
	`, match.Id, match.Sport, match.Date, match.ParticipantNber, match.CurrentState, match.Score1, match.Score2, match.CourtID, match.CreatorID, match.CreatedAt, now)

	return err
}

func (db Database) CountUsersByMatchAndTeam(ctx context.Context, matchId string, team int) (int, error) {
	var count int
	err := db.Database.GetContext(ctx, &count, `
        SELECT COUNT(*) 
        FROM user_match 
        WHERE match_id = $1 AND team = $2`, matchId, team)

	if err != nil {
		return 0, fmt.Errorf("échec du comptage des utilisateurs pour le param %s et len'équipe %d : %w", matchId, team, err)
	}

	return count, nil
}

func (db Database) CountUsersByMatch(ctx context.Context, matchId string) (int, error) {
	var count int
	err := db.Database.GetContext(ctx, &count, `
        SELECT COUNT(*) 
        FROM user_match 
        WHERE match_id = $1`, matchId)

	if err != nil {
		return 0, fmt.Errorf("échec du comptage des utilisateurs pour le param %s : %w", matchId, err)
	}

	return count, nil
}

func (db Database) GetUserInMatch(ctx context.Context, userID, matchID string) (*models.DBUserMatch, error) {
	var um models.DBUserMatch
	err := db.Database.GetContext(ctx, &um, `
		SELECT user_id, match_id, team, created_at
		FROM user_match
		WHERE user_id = $1 AND match_id = $2
		LIMIT 1
	`, userID, matchID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying user_match: %w", err)
	}
	return &um, nil
}

func (db Database) GetUserMatchesByMatchID(ctx context.Context, matchID string) ([]models.DBUserMatch, error) {
	var rows []models.DBUserMatch
	err := db.Database.SelectContext(ctx, &rows, `
		SELECT user_id, match_id, team, created_at
		FROM user_match
		WHERE match_id = $1
	`, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user_match rows: %w", err)
	}
	return rows, nil
}

func (db Database) GetUserWinrate(ctx context.Context, userID string) (*int, error) {
	type winrateRow struct {
		Wins  int `db:"wins"`
		Total int `db:"total"`
	}

	var row winrateRow
	err := db.Database.GetContext(ctx, &row, `
        SELECT
          COALESCE(SUM(
            CASE
              WHEN (um.team = 1 AND m.score1 > m.score2) OR
                   (um.team = 2 AND m.score2 > m.score1)
              THEN 1 ELSE 0
            END
          ), 0) AS wins,
          COUNT(*) AS total
        FROM user_match um
        JOIN matches m ON m.id = um.match_id
        WHERE um.user_id = $1
          AND m.current_state = 'Termine'
          AND m.score1 IS NOT NULL
          AND m.score2 IS NOT NULL
          AND m.score1 <> m.score2
    `, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to compute winrate: %w", err)
	}

	if row.Total == 0 {
		return nil, nil
	}

	pct := int((float64(row.Wins)/float64(row.Total))*100.0 + 0.5)
	return &pct, nil
}

func (db Database) GetUsersByMatchIDs(ctx context.Context, matchIDs []string) (map[string][]models.DBUsers, error) {
	query := `
        SELECT mu.match_id,
               u.id, u.username, u.email, u.bio, u.current_field_id, 
               u.password, u.created_at, u.updated_at
        FROM user_match mu
        JOIN users u ON u.id = mu.user_id
        WHERE mu.match_id = ANY($1)
    `

	type row struct {
		MatchID string `db:"match_id"`
		models.DBUsers
	}

	var rows []row
	if err := db.Database.SelectContext(ctx, &rows, query, matchIDs); err != nil {
		return nil, fmt.Errorf("failed to fetch users by matchIDs: %w", err)
	}

	result := make(map[string][]models.DBUsers)
	for _, r := range rows {
		result[r.MatchID] = append(result[r.MatchID], r.DBUsers)
	}

	return result, nil
}
