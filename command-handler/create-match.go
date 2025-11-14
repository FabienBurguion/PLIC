package main

import (
	"PLIC/database"
	"PLIC/models"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type CreateMatchOptions struct {
	UserID       string
	CourtID      string
	Sport        string
	Participants int
	Team         int
}

func RunCreateMatch(ctx context.Context, db database.Database, opt CreateMatchOptions) error {
	if opt.Sport == "" {
		opt.Sport = string(models.Basket)
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

	matchID := uuid.NewString()
	now := time.Now()

	match := models.DBMatches{
		Id:              matchID,
		Sport:           models.Sport(opt.Sport),
		Date:            now,
		ParticipantNber: opt.Participants,
		CurrentState:    models.ManqueJoueur,
		Score1:          nil,
		Score2:          nil,
		CreatorID:       opt.UserID,
		CourtID:         opt.CourtID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	log.Info().
		Str("match_id", matchID).
		Str("user_id", opt.UserID).
		Str("court_id", opt.CourtID).
		Str("sport", opt.Sport).
		Int("participants", opt.Participants).
		Int("team", opt.Team).
		Msg("🔧 create-match: création du match")

	if err := db.CreateMatch(ctx, match); err != nil {
		return fmt.Errorf("insert match: %w", err)
	}

	um := models.DBUserMatch{
		UserID:    opt.UserID,
		MatchID:   matchID,
		Team:      opt.Team,
		CreatedAt: now,
	}

	if err := db.CreateUserMatch(ctx, um); err != nil {
		_ = db.DeleteMatch(ctx, matchID)
		return fmt.Errorf("insert user_match: %w", err)
	}

	log.Info().Str("match_id", matchID).Msg("🟢 Match créé et créateur inscrit")
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
