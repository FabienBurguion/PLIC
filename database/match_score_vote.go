package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (db Database) GetMatchScoreVote(ctx context.Context, matchID, userID string) (*models.DBMatchScoreVote, error) {
	var v models.DBMatchScoreVote
	err := db.Database.GetContext(ctx, &v, `
		SELECT match_id, user_id, team, score1, score2, created_at
		FROM match_score_vote
		WHERE match_id = $1 AND user_id = $2
	`, matchID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch match score vote: %w", err)
	}
	return &v, nil
}

func (db Database) UpsertMatchScoreVote(ctx context.Context, matchScoreVote models.DBMatchScoreVote) error {
	_, err := db.Database.ExecContext(ctx, `
		INSERT INTO match_score_vote (match_id, user_id, team, score1, score2)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (match_id, user_id)
		DO UPDATE SET score1 = EXCLUDED.score1,
					  score2 = EXCLUDED.score2,
					  team   = EXCLUDED.team,
					  created_at = NOW();
	`, matchScoreVote.MatchId, matchScoreVote.UserId, matchScoreVote.Team, matchScoreVote.Score1, matchScoreVote.Score2)
	return err
}

func (db Database) HasConsensusScore(ctx context.Context, matchID string, team, score1, score2 int) (bool, error) {
	var exists bool
	err := db.Database.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1
			FROM match_score_vote
			WHERE match_id = $1
			  AND team <> $2
			  AND score1 = $3
			  AND score2 = $4
		)
	`, matchID, team, score1, score2)
	if err != nil {
		return false, fmt.Errorf("failed to check consensus score: %w", err)
	}
	return exists, nil
}

func (db Database) HasOtherTeamVote(ctx context.Context, matchID string, team int, userID string) (bool, error) {
	var exists bool
	err := db.Database.GetContext(ctx, &exists, `
		SELECT EXISTS (
		  SELECT 1
		  FROM match_score_vote
		  WHERE match_id = $1
		    AND team = $2
		    AND user_id <> $3
		)
	`, matchID, team, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check team duplicate vote: %w", err)
	}
	return exists, nil
}

func (db Database) GetScoreVoteByMatchAndTeam(ctx context.Context, matchID string, team int) (*models.DBMatchScoreVote, error) {
	var row models.DBMatchScoreVote
	err := db.Database.GetContext(ctx, &row, `
		SELECT match_id, user_id, team, score1, score2, created_at
		FROM match_score_vote
		WHERE match_id = $1 AND team = $2
		LIMIT 1
	`, matchID, team)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch score vote for match %s team %d: %w", matchID, team, err)
	}
	return &row, nil
}
