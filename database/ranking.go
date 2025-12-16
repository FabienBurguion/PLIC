package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (db Database) GetRankedFieldsByUserID(ctx context.Context, userID string) ([]models.Field, error) {
	var fields []models.Field

	const q = `
	WITH user_court_sports AS (
	  SELECT DISTINCT m.court_id, m.sport
	  FROM matches m
	  JOIN user_match um ON um.match_id = m.id
	  WHERE um.user_id = $1
	),
	ranked AS (
	  SELECT
		RANK() OVER (
		  PARTITION BY r.court_id, r.sport
		  ORDER BY r.elo DESC
		) AS ranking,
		r.user_id,
		c.name,
		r.elo,
		r.sport
	  FROM ranking r
	  JOIN courts c ON c.id = r.court_id
	  JOIN user_court_sports ucs
		ON ucs.court_id = r.court_id
	   AND ucs.sport    = r.sport
	)
	SELECT ranking, name, elo, sport
	FROM ranked
	WHERE user_id = $1
	ORDER BY name, sport;
`
	if err := db.Database.SelectContext(ctx, &fields, q, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching ranked fields: %w", err)
	}
	return fields, nil
}

func (db Database) InsertRanking(ctx context.Context, ranking models.DBRanking) error {
	_, err := db.Database.ExecContext(ctx, `
		INSERT INTO ranking (user_id, court_id, elo, sport, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, court_id, sport) DO UPDATE
		SET elo = EXCLUDED.elo,
		    updated_at = EXCLUDED.updated_at`,
		ranking.UserID, ranking.CourtID, ranking.Elo, ranking.Sport, ranking.CreatedAt, ranking.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error inserting ranking: %w", err)
	}
	return nil
}

func (db Database) GetRankingByUserAndCourt(ctx context.Context, userID, courtID string) (*models.DBRanking, error) {
	var ranking models.DBRanking
	err := db.Database.GetContext(ctx, &ranking,
		`SELECT user_id, court_id, elo, created_at, updated_at
		FROM ranking
		WHERE user_id = $1 AND court_id = $2`,
		userID, courtID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch ranking: %w", err)
	}
	return &ranking, nil
}

func (db Database) GetRankingsByCourtID(ctx context.Context, courtID string, sport models.Sport) ([]models.DBRanking, error) {
	var rows []models.DBRanking
	err := db.Database.SelectContext(ctx, &rows, `
				SELECT user_id, court_id, elo, sport, created_at, updated_at
				FROM ranking
				WHERE court_id = $1
				AND sport = $2
				ORDER BY elo , user_id
			`, courtID, sport)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rankings: %w", err)
	}
	return rows, nil
}

func (db Database) GetRankingByUserCourtSport(ctx context.Context, userID, courtID string, sport models.Sport) (*models.DBRanking, error) {
	var ranking models.DBRanking
	err := db.Database.GetContext(ctx, &ranking, `
		SELECT user_id, court_id, sport, elo, created_at, updated_at
		FROM ranking
		WHERE user_id = $1 AND court_id = $2 AND sport = $3
		LIMIT 1`,
		userID, courtID, sport,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch ranking: %w", err)
	}
	return &ranking, nil
}
