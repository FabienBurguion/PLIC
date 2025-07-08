package database

import (
	"PLIC/models"
	"context"
	"fmt"
)

func (db Database) GetRankedFieldsByUserID(ctx context.Context, userID string) ([]models.Field, error) {
	var fields []models.Field
	err := db.Database.SelectContext(ctx, &fields, `
		SELECT
			RANK() OVER (ORDER BY r.elo DESC) AS ranking,
			c.name,
			r.elo AS score
		FROM ranking r
		JOIN courts c ON c.id = r.court_id
		WHERE r.user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching ranked fields: %w", err)
	}
	return fields, nil
}

func (db Database) InsertRanking(ctx context.Context, ranking models.DBRanking) error {
	_, err := db.Database.ExecContext(ctx, `
		INSERT INTO ranking (user_id, court_id, elo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, court_id) DO UPDATE SET elo = $3`,
		ranking.UserID, ranking.CourtID, ranking.Elo, ranking.CreatedAt, ranking.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error inserting ranking: %w", err)
	}
	return nil
}
