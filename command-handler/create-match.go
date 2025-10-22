package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type CreateMatchOptions struct {
	MatchID      string
	UserID       string
	CourtID      string
	Sport        string
	Participants int
	Team         int
}

func RunCreateMatch(ctx context.Context, db *sqlx.DB, opt CreateMatchOptions) error {
	if opt.Sport == "" {
		opt.Sport = "basket"
	}
	if !isValidSport(opt.Sport) {
		return fmt.Errorf("sport invalide: %s (attendu: basket|foot|ping-pong)", opt.Sport)
	}
	if opt.Team != 1 && opt.Team != 2 {
		return fmt.Errorf("team invalide: %d (attendu: 1 ou 2)", opt.Team)
	}
	if opt.Participants <= 0 {
		opt.Participants = 2
	}

	if strings.TrimSpace(opt.MatchID) == "" {
		opt.MatchID = uuid.NewString()
	}

	log.Info().
		Str("match_id", opt.MatchID).
		Str("user_id", opt.UserID).
		Str("court_id", opt.CourtID).
		Str("sport", opt.Sport).
		Int("participants", opt.Participants).
		Int("team", opt.Team).
		Msg("🔧 create-match: paramètres")

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const insertMatch = `
INSERT INTO matches (id, sport, date, participant_nber, current_state, court_id, created_at, updated_at)
VALUES ($1, $2, NOW(), $3, 'Manque joueur', $4, NOW(), NOW());
`
	if _, err := tx.ExecContext(ctx, insertMatch, opt.MatchID, opt.Sport, opt.Participants, opt.CourtID); err != nil {
		return fmt.Errorf("insert matches: %w", err)
	}

	const insertUserMatch = `
INSERT INTO user_match (user_id, match_id, team, created_at)
VALUES ($1, $2, $3, $4);
`
	if _, err := tx.ExecContext(ctx, insertUserMatch, opt.UserID, opt.MatchID, opt.Team, time.Now()); err != nil {
		return fmt.Errorf("insert user_match: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Info().Str("match_id", opt.MatchID).Msg("🟢 Match créé et créateur inscrit")
	return nil
}

func isValidSport(s string) bool {
	switch strings.ToLower(s) {
	case "basket", "foot", "ping-pong":
		return true
	default:
		return false
	}
}

func ensureRowExists(ctx context.Context, tx *sqlx.Tx, query string, args ...any) error {
	var exists bool
	if err := tx.QueryRowxContext(ctx, query, args...).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return errors.New("not found")
	}
	return nil
}
