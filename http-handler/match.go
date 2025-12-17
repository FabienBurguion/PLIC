package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func (s *Service) buildMatchesResponse(ctx context.Context, matches []models.DBMatches) []models.MatchResponse {
	logger := log.With().
		Str("method", "buildMatchesResponse").
		Int("match_count", len(matches)).
		Logger()

	logger.Info().Msg("building matches batch")

	if len(matches) == 0 {
		return nil
	}

	matchIDs := make([]string, 0, len(matches))
	courtIDs := make(map[string]struct{})
	for _, m := range matches {
		matchIDs = append(matchIDs, m.Id)
		courtIDs[m.CourtID] = struct{}{}
	}

	usersByMatch, err := s.db.GetUsersByMatchIDs(ctx, matchIDs)
	if err != nil {
		logger.Error().Err(err).Msg("prefetching users failed")
		return nil
	}

	courtIDsSlice := make([]string, 0, len(courtIDs))
	for id := range courtIDs {
		courtIDsSlice = append(courtIDsSlice, id)
	}
	courts, err := s.db.GetCourtsByIDs(ctx, courtIDsSlice)
	if err != nil {
		logger.Error().Err(err).Msg("prefetching courts failed")
		return nil
	}

	courtMap := make(map[string]models.DBCourt, len(courts))
	for _, c := range courts {
		courtMap[c.Id] = c
	}

	userIDSet := make(map[string]struct{})
	for _, us := range usersByMatch {
		for _, u := range us {
			userIDSet[u.Id] = struct{}{}
		}
	}
	if len(userIDSet) == 0 {
		logger.Warn().Msg("no users found for matches")
	}

	userIDs := make([]string, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	statsByUser, err := s.db.GetUserStatsByIDs(ctx, userIDs)
	if err != nil {
		logger.Error().Err(err).Msg("prefetching user stats failed")
	}

	profilePics := make(map[string]string, len(userIDs))
	var mu sync.Mutex

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, uid := range userIDs {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			pic, err := s.s3Service.GetProfilePicture(ctx, userID)
			if err != nil {
				logger.Warn().Err(err).Str("user_id", userID).Msg("failed to get profile picture")
				return
			}

			mu.Lock()
			profilePics[userID] = pic.URL
			mu.Unlock()
		}(uid)
	}
	wg.Wait()

	responses := make([]models.MatchResponse, 0, len(matches))
	for _, match := range matches {
		users := usersByMatch[match.Id]

		court, ok := courtMap[match.CourtID]
		if !ok {
			logger.Warn().
				Str("match_id", match.Id).
				Str("court_id", match.CourtID).
				Msg("no court found for match")
			continue
		}

		userResponses := make([]models.UserResponse, len(users))
		for i, u := range users {
			picURL := profilePics[u.Id]
			userResponses[i] = s.buildUserResponseFast(&u, picURL, statsByUser[u.Id])
		}

		responses = append(responses, models.MatchResponse{
			Id:              match.Id,
			Sport:           match.Sport,
			Place:           court.Name,
			Date:            match.Date,
			NbreParticipant: match.ParticipantNber,
			CurrentState:    match.CurrentState,
			Score1:          match.Score1,
			Score2:          match.Score2,
			Users:           userResponses,
			CreatedAt:       match.CreatedAt,
		})
	}

	logger.Info().Int("responses_built", len(responses)).Msg("matches built successfully")
	return responses
}

// GetMatchByID godoc
// @Summary      Récupère un match par son ID
// @Description  Retourne les informations d’un match en fonction de son identifiant passé en paramètre de requête
// @Tags         match
// @Produce      json
// @Param        id   path      string  true  "Identifiant du match"
// @Success      200  {object}  models.MatchResponse "Match trouvé"
// @Failure      400  {object}  models.Error         "ID manquant ou invalide"
// @Failure      401   {object}  models.Error       "Utilisateur non autorisé"
// @Failure      404  {object}  models.Error         "Match non trouvé"
// @Failure      500  {object}  models.Error         "Erreur serveur ou base de données"
// @Router       /match/{id} [get]
func (s *Service) GetMatchByID(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "GetMatchByID").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	id := chi.URLParam(r, "id")
	logger := baseLogger.With().Str("match_id", id).Logger()

	if id == "" {
		logger.Warn().Msg("missing id in url params")
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	ctx := r.Context()
	match, err := s.db.GetMatchById(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("db error fetching match")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if match == nil {
		logger.Warn().Msg("match not found")
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}

	responses := s.buildMatchesResponse(ctx, []models.DBMatches{*match})
	if len(responses) == 0 {
		logger.Error().Msg("failed to build match response")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to build match response")
	}

	logger.Info().Msg("match fetched successfully")
	return httpx.Write(w, http.StatusOK, responses[0])
}

// GetMatchesByUserID godoc
// @Summary      Liste des matchs d’un utilisateur
// @Description  Retourne les matchs auxquels un utilisateur a participé
// @Tags         match
// @Produce      json
// @Param        userId   path      string  true  "Identifiant de l'utilisateur"
// @Success      200  {array}   models.MatchResponse
// @Failure      400  {object}  models.Error
// @Failure      401  {object}  models.Error
// @Failure      404  {object}  models.Error
// @Failure      500  {object}  models.Error
// @Router       /user/matches/{userId} [get]
func (s *Service) GetMatchesByUserID(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	logger := log.With().
		Str("method", "GetMatchesByUserID").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		logger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	if ai.UserID == "" {
		logger.Warn().Msg("missing userId token")
		return httpx.WriteError(w, http.StatusBadRequest, "missing userId in token")
	}

	ctx := r.Context()

	dbMatches, err := s.db.GetMatchesByUserID(ctx, ai.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("db get matches by user failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}

	res := s.buildMatchesResponse(ctx, dbMatches)
	logger.Info().Int("count", len(res)).Msg("matches fetched for user")
	return httpx.Write(w, http.StatusOK, res)
}

// GetMatchesByCourtId godoc
// @Summary      Liste des matchs pour un court
// @Description  Retourne les matchs associés à un terrain (court) via son ID
// @Tags         match
// @Produce      json
// @Param        courtId   path      string  true  "Identifiant du terrain"
// @Success      200  {array}   models.MatchResponse
// @Failure      400  {object}  models.Error  "ID manquant"
// @Failure      401  {object}  models.Error  "Utilisateur non autorisé"
// @Failure      404  {object}  models.Error  "Aucun match trouvé pour ce terrain"
// @Failure      500  {object}  models.Error  "Erreur interne serveur ou base"
// @Router       /matches/court/{courtId} [get]
func (s *Service) GetMatchesByCourtId(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "GetMatchesByCourtId").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	courtID := chi.URLParam(r, "courtId")
	logger := baseLogger.With().Str("court_id", courtID).Logger()

	if courtID == "" {
		logger.Warn().Msg("missing courtId in url params")
		return httpx.WriteError(w, http.StatusBadRequest, "missing courtId in url params")
	}

	ctx := r.Context()

	matches, err := s.db.GetMatchesByCourtId(ctx, courtID)
	if err != nil {
		logger.Error().Err(err).Msg("db get matches by court failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}

	res := s.buildMatchesResponse(ctx, matches)
	logger.Info().Int("count", len(res)).Msg("matches fetched for court")
	return httpx.Write(w, http.StatusOK, res)
}

// GetAllMatches godoc
// @Summary      Liste tous les matchs
// @Description  Retourne la liste complète de tous les matchs stockés en base
// @Tags         match
// @Produce      json
// @Success      200  {array}   models.MatchResponse "Liste des matchs"
// @Failure      401   {object}  models.Error       "Utilisateur non autorisé"
// @Failure      500  {object}  models.Error          "Erreur serveur lors de la récupération des matchs"
// @Router       /match/all [get]
func (s *Service) GetAllMatches(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "GetAllMatches").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()

	matches, err := s.db.GetAllMatches(ctx)
	if err != nil {
		baseLogger.Error().Err(err).Msg("db get all matches failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch matches")
	}

	res := s.buildMatchesResponse(ctx, matches)
	baseLogger.Info().Int("count", len(res)).Msg("all matches fetched")
	return httpx.Write(w, http.StatusOK, res)
}

// CreateMatch godoc
// @Summary      Crée un nouveau match
// @Description  Enregistre un nouveau match en base de données à partir des données fournies en JSON
// @Tags         match
// @Accept       json
// @Produce      json
// @Param        match  body      models.MatchRequest  true  "Objet match à créer"
// @Success      201    {object}  map[string]string    "Match créé avec succès"
// @Failure      400    {object}  models.Error         "Données invalides ou champ ID manquant"
// @Failure      401   {object}  models.Error       "Utilisateur non autorisé"
// @Failure      500    {object}  models.Error         "Erreur lors de la création du match"
// @Router       /match [post]
func (s *Service) CreateMatch(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "CreateMatch").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	var match models.MatchRequest
	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) { _ = Body.Close() }(r.Body)
	if err := decoder.Decode(&match); err != nil {
		baseLogger.Warn().Err(err).Msg("invalid JSON body")
		return httpx.WriteError(w, http.StatusBadRequest, "invalid JSON")
	}

	logger := baseLogger.With().
		Str("court_id", match.CourtID).
		Int("nbre_participant", match.NbreParticipant).
		Str("sport", string(match.Sport)).
		Logger()

	if match.NbreParticipant < 2 || match.NbreParticipant%2 != 0 {
		logger.Warn().Msg("invalid number of participant")
		return httpx.WriteError(w, http.StatusBadRequest, "invalid number of participant")
	}

	ctx := r.Context()

	court, err := s.db.GetCourtByID(ctx, match.CourtID)
	if err != nil {
		logger.Error().Err(err).Msg("db get court failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch court")
	}
	if court == nil {
		logger.Warn().Msg("court not found")
		return httpx.WriteError(w, http.StatusBadRequest, "court not found")
	}

	matchDb := match.ToDBMatches(s.clock.Now(), ai.UserID)

	if err := s.db.CreateMatch(ctx, matchDb); err != nil {
		logger.Error().Err(err).Msg("db create match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to create match")
	}

	existing, err := s.db.GetRankingByUserCourtSport(ctx, ai.UserID, match.CourtID, match.Sport)
	if err != nil {
		logger.Error().Err(err).Msg("db get ranking failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check ranking")
	}
	if existing == nil {
		if err := s.db.InsertRanking(ctx, models.DBRanking{
			UserID:    ai.UserID,
			CourtID:   match.CourtID,
			Sport:     match.Sport,
			Elo:       DefaultElo,
			CreatedAt: s.clock.Now(),
			UpdatedAt: s.clock.Now(),
		}); err != nil {
			logger.Error().Err(err).Msg("db insert default ranking failed")
			return httpx.WriteError(w, http.StatusInternalServerError, "failed to create default ranking")
		}
	}

	if err := s.db.CreateUserMatch(ctx, models.DBUserMatch{
		UserID:    ai.UserID,
		MatchID:   matchDb.Id,
		Team:      1, // creator joins team 1
		CreatedAt: s.clock.Now(),
	}); err != nil {
		logger.Error().Err(err).Msg("db create user_match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to associate user to match")
	}

	logger.Info().Str("match_id", matchDb.Id).Msg("match created")
	return httpx.Write(w, http.StatusCreated, models.CreateMatchResponse{Id: matchDb.Id})
}

// JoinMatch godoc
// @Summary      Un utilisateur rejoint un match
// @Description  Permet à un utilisateur authentifié de rejoindre un match existant, si ce n’est pas déjà fait
// @Tags         match
// @Produce      json
// @Param        id    path      string             true  "Identifiant du match"
// @Param        body  body      models.JoinMatchRequest  true   "Informations pour rejoindre un match (team)"
// @Success      200
// @Failure      400   {object}  models.Error       "Identifiant manquant"
// @Failure      401   {object}  models.Error       "Utilisateur non autorisé"
// @Failure      404   {object}  models.Error       "Match non trouvé"
// @Failure      409   {object}  models.Error       "Utilisateur déjà inscrit au match"
// @Failure      500   {object}  models.Error       "Erreur lors de l'inscription de l'utilisateur au match"
// @Router       /join/match/{id} [post]
func (s *Service) JoinMatch(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "JoinMatch").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()

	matchID := chi.URLParam(r, "id")
	logger := baseLogger.With().Str("match_id", matchID).Logger()

	if matchID == "" {
		logger.Warn().Msg("missing match ID")
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	var matchRequest models.JoinMatchRequest
	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) { _ = Body.Close() }(r.Body)
	if err := decoder.Decode(&matchRequest); err != nil {
		logger.Warn().Err(err).Msg("invalid JSON body")
		return httpx.WriteError(w, http.StatusBadRequest, "invalid JSON")
	}

	match, err := s.db.GetMatchById(ctx, matchID)
	if err != nil {
		logger.Error().Err(err).Msg("db get match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch match")
	}
	if match == nil {
		logger.Warn().Msg("match not found")
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}

	if match.CurrentState != models.ManqueJoueur {
		logger.Warn().Str("state", string(match.CurrentState)).Msg("match not in ManqueJoueur")
		return httpx.WriteError(w, http.StatusBadRequest, "match is not in the right state")
	}

	exists, err := s.db.IsUserInMatch(ctx, ai.UserID, matchID)
	if err != nil {
		logger.Error().Err(err).Msg("db check user in match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check user in match")
	}
	if exists {
		logger.Warn().Msg("user already in match")
		return httpx.WriteError(w, http.StatusConflict, "user already joined the match")
	}

	count, err := s.db.CountUsersByMatchAndTeam(ctx, matchID, matchRequest.Team)
	if err != nil {
		logger.Error().Err(err).Msg("db count users by match/team failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to count users by match and team")
	}
	if count >= match.ParticipantNber/2 {
		logger.Warn().Int("team", matchRequest.Team).Msg("team full")
		return httpx.WriteError(w, http.StatusBadRequest, "this team is full")
	}

	existing, err := s.db.GetRankingByUserCourtSport(ctx, ai.UserID, match.CourtID, match.Sport)
	if err != nil {
		logger.Error().Err(err).Msg("db get ranking failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check ranking")
	}
	if existing == nil {
		if err := s.db.InsertRanking(ctx, models.DBRanking{
			UserID:    ai.UserID,
			CourtID:   match.CourtID,
			Elo:       DefaultElo,
			Sport:     match.Sport,
			CreatedAt: s.clock.Now(),
			UpdatedAt: s.clock.Now(),
		}); err != nil {
			logger.Error().Err(err).Msg("db insert default ranking failed")
			return httpx.WriteError(w, http.StatusInternalServerError, "failed to create default ranking")
		}
	}

	if err := s.db.CreateUserMatch(ctx, models.DBUserMatch{
		UserID:    ai.UserID,
		MatchID:   matchID,
		Team:      matchRequest.Team,
		CreatedAt: s.clock.Now(),
	}); err != nil {
		logger.Error().Err(err).Msg("db create user_match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to join match")
	}

	newCount, err := s.db.CountUsersByMatch(ctx, matchID)
	if err != nil {
		logger.Error().Err(err).Msg("db count users by match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to count users by match")
	}
	if newCount == match.ParticipantNber {
		match.CurrentState = models.Valide
	}

	match.UpdatedAt = s.clock.Now()
	if err := s.db.UpsertMatch(ctx, *match, s.clock.Now()); err != nil {
		logger.Error().Err(err).Msg("db upsert match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to update match")
	}

	logger.Info().Msg("user joined match successfully")
	return httpx.Write(w, http.StatusOK, nil)
}

// DeleteMatch godoc
// @Summary      Supprime un match
// @Description  Supprime un match via son ID
// @Tags         match
// @Produce      json
// @Param        id   path      string  true  "Identifiant du match à supprimer"
// @Success      200
// @Failure      400  {object}  models.Error      "ID manquant"
// @Failure      401  {object}  models.Error      "Utilisateur non autorisé"
// @Failure      500  {object}  models.Error      "Erreur lors de la suppression du match"
// @Router       /match/{id} [delete]
func (s *Service) DeleteMatch(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "DeleteMatch").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	matchID := chi.URLParam(r, "id")
	logger := baseLogger.With().Str("match_id", matchID).Logger()

	if matchID == "" {
		logger.Warn().Msg("missing match ID")
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	ctx := r.Context()
	if err := s.db.DeleteMatch(ctx, matchID); err != nil {
		logger.Error().Err(err).Msg("db delete match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to delete match")
	}

	logger.Info().Msg("match deleted")
	return httpx.Write(w, http.StatusOK, nil)
}

func (s *Service) applyEloForMatch(ctx context.Context, match models.DBMatches, score1, score2 int) error {
	userMatches, err := s.db.GetUserMatchesByMatchID(ctx, match.Id)
	if err != nil {
		return err
	}
	if len(userMatches) == 0 {
		return nil
	}

	var team1Users, team2Users []string
	for _, um := range userMatches {
		switch um.Team {
		case 1:
			team1Users = append(team1Users, um.UserID)
		case 2:
			team2Users = append(team2Users, um.UserID)
		}
	}
	if len(team1Users) == 0 || len(team2Users) == 0 {
		return nil
	}

	now := s.clock.Now()

	getRanks := func(userIDs []string) ([]models.DBRanking, error) {
		out := make([]models.DBRanking, 0, len(userIDs))
		for _, uid := range userIDs {
			rk, err := s.db.GetRankingByUserCourtSport(ctx, uid, match.CourtID, match.Sport)
			if err != nil {
				return nil, err
			}
			if rk == nil {
				return nil, fmt.Errorf("ranking missing for user=%s court=%s sport=%s", uid, match.CourtID, match.Sport)
			}
			out = append(out, *rk)
		}
		return out, nil
	}

	r1, err := getRanks(team1Users)
	if err != nil {
		return err
	}
	r2, err := getRanks(team2Users)
	if err != nil {
		return err
	}

	avg := func(rs []models.DBRanking) float64 {
		if len(rs) == 0 {
			return float64(DefaultElo)
		}
		sum := 0
		for _, r := range rs {
			sum += r.Elo
		}
		return float64(sum) / float64(len(rs))
	}

	rTeam1 := avg(r1)
	rTeam2 := avg(r2)

	var s1, s2 float64
	switch {
	case score1 > score2:
		s1, s2 = 1.0, 0.0
	case score1 < score2:
		s1, s2 = 0.0, 1.0
	default:
		s1, s2 = 0.5, 0.5
	}

	exp := func(rA, rB float64) float64 {
		return 1.0 / (1.0 + math.Pow(10, (rB-rA)/400.0))
	}
	e1 := exp(rTeam1, rTeam2)
	e2 := exp(rTeam2, rTeam1)

	applyDelta := func(rs []models.DBRanking, S, E float64) []models.DBRanking {
		out := make([]models.DBRanking, len(rs))
		for i, rk := range rs {
			delta := int(math.Round(float64(KFactor) * (S - E)))
			rk.Elo = rk.Elo + delta
			rk.UpdatedAt = now
			out[i] = rk
		}
		return out
	}

	r1New := applyDelta(r1, s1, e1)
	r2New := applyDelta(r2, s2, e2)

	for _, rk := range append(r1New, r2New...) {
		if err := s.db.InsertRanking(ctx, rk); err != nil {
			return err
		}
	}
	return nil
}

// UpdateMatchScore godoc
// @Summary      Met à jour le score d’un match
// @Description  Met à jour les scores (score1 et score2) d’un match via son ID
// @Tags         match
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "ID du match"
// @Param        body  body      models.UpdateScoreRequest true  "Nouveaux scores"
// @Success      200
// @Failure      400   {object}  models.Error
// @Failure      401   {object}  models.Error
// @Failure      404   {object}  models.Error
// @Failure      500   {object}  models.Error
// @Router       /score/match/{id} [patch]
func (s *Service) UpdateMatchScore(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "UpdateMatchScore").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()
	id := chi.URLParam(r, "id")
	logger := baseLogger.With().Str("match_id", id).Logger()

	if id == "" {
		logger.Warn().Msg("missing match ID")
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	var req models.UpdateScoreRequest
	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) { _ = Body.Close() }(r.Body)
	if err := decoder.Decode(&req); err != nil {
		logger.Warn().Err(err).Msg("invalid JSON body")
		return httpx.WriteError(w, http.StatusBadRequest, "invalid JSON")
	}

	match, err := s.db.GetMatchById(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("db get match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if match == nil {
		logger.Warn().Msg("match not found")
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}

	if match.CurrentState != models.ManqueScore {
		logger.Warn().Str("state", string(match.CurrentState)).Msg("match not in ManqueScore")
		return httpx.WriteError(w, http.StatusBadRequest, "match is not in the right state")
	}

	userMatch, err := s.db.GetUserInMatch(ctx, ai.UserID, id)
	if err != nil {
		logger.Error().Err(err).Msg("db get user in match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if userMatch == nil {
		logger.Warn().Msg("user not in match")
		return httpx.WriteError(w, http.StatusNotFound, "user in this match not found")
	}

	hasSameTeamOtherVote, err := s.db.HasOtherTeamVote(ctx, id, userMatch.Team, ai.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("db check other team vote failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check team vote")
	}
	if hasSameTeamOtherVote {
		logger.Warn().Msg("team already has a vote")
		return httpx.WriteError(w, http.StatusBadRequest, "this team already has a vote")
	}

	if err := s.db.UpsertMatchScoreVote(ctx, models.DBMatchScoreVote{
		MatchId: id,
		UserId:  ai.UserID,
		Team:    userMatch.Team,
		Score1:  req.Score1,
		Score2:  req.Score2,
	}); err != nil {
		logger.Error().Err(err).Msg("db upsert score vote failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to upsert score vote")
	}

	hasConsensus, err := s.db.HasConsensusScore(ctx, id, userMatch.Team, req.Score1, req.Score2)
	if err != nil {
		logger.Error().Err(err).Msg("db check consensus failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check consensus")
	}

	if hasConsensus {
		match.CurrentState = models.Termine
		if err := s.applyEloForMatch(ctx, *match, req.Score1, req.Score2); err != nil {
			logger.Error().Err(err).Msg("apply ELO failed")
			return httpx.WriteError(w, http.StatusInternalServerError, "failed to update rankings")
		}

		court, err := s.db.GetCourtByID(ctx, match.CourtID)
		if err != nil || court == nil {
			logger.Error().Err(err).Msg("db get court by id failed (email for result mail)")
		} else {
			userMatches, err := s.db.GetUserMatchesByMatchID(ctx, id)
			if err != nil {
				logger.Error().Err(err).Msg("db get user_matches failed (email for result mail)")
			} else {
				for _, um := range userMatches {
					u, err := s.db.GetUserById(ctx, um.UserID)
					if err != nil || u == nil {
						logger.Error().Err(err).Str("user_id", um.UserID).Msg("db get user by id failed (email for result mail)")
						continue
					}

					teamScore, oppScore := req.Score1, req.Score2
					if um.Team == 2 {
						teamScore, oppScore = req.Score2, req.Score1
					}

					if err := s.mailer.SendMatchResultEmail(id, u.Email, u.Username, match.Sport, court.Name, teamScore, oppScore); err != nil {
						logger.Error().
							Err(err).
							Str("email", u.Email).
							Int("team", um.Team).
							Int("team_score", teamScore).
							Int("opp_score", oppScore).
							Msg("sending match result email failed")
					} else {
						logger.Info().
							Str("email", u.Email).
							Int("team", um.Team).
							Int("team_score", teamScore).
							Int("opp_score", oppScore).
							Msg("match result email sent")
					}
				}
			}
		}
	}

	match.Score1 = &req.Score1
	match.Score2 = &req.Score2
	match.UpdatedAt = s.clock.Now()

	if err := s.db.UpsertMatch(ctx, *match, s.clock.Now()); err != nil {
		logger.Error().Err(err).Msg("db upsert match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to update match")
	}

	logger.Info().Msg("match score updated")
	return httpx.Write(w, http.StatusOK, nil)
}

// StartMatch godoc
// @Summary      Démarre un match
// @Description  Passe un match de l’état "Valide" à "En cours" et met à jour la date de début à maintenant.
// @Tags         match
// @Produce      json
// @Param        id    path      string  true  "ID du match"
// @Success      200
// @Failure      400   {object}  models.Error  "ID manquant, mauvais état, ou utilisateur non inscrit au match"
// @Failure      401   {object}  models.Error  "Utilisateur non autorisé"
// @Failure      404   {object}  models.Error  "Match non trouvé"
// @Failure      500   {object}  models.Error  "Erreur serveur ou base de données"
// @Router       /match/{id}/start [patch]
func (s *Service) StartMatch(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "StartMatch").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()
	id := chi.URLParam(r, "id")
	logger := baseLogger.With().Str("match_id", id).Logger()

	if id == "" {
		logger.Warn().Msg("missing match ID")
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	match, err := s.db.GetMatchById(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("db get match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if match == nil {
		logger.Warn().Msg("match not found")
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}
	if match.CurrentState != models.Valide {
		logger.Warn().Str("state", string(match.CurrentState)).Msg("match not in Valide")
		return httpx.WriteError(w, http.StatusBadRequest, "match is not in the right state")
	}
	if match.CreatorID != ai.UserID {
		msg := "user is not the match creator"
		logger.Warn().Str("creator_id", match.CreatorID).Msg(msg)
		return httpx.WriteError(w, http.StatusForbidden, msg)
	}

	userInMatch, err := s.db.IsUserInMatch(ctx, ai.UserID, id)
	if err != nil {
		logger.Error().Err(err).Msg("db check user in match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if !userInMatch {
		logger.Warn().Msg("user not in match")
		return httpx.WriteError(w, http.StatusBadRequest, "user is not in the match")
	}

	match.Date = s.clock.Now()
	match.CurrentState = models.EnCours
	match.UpdatedAt = s.clock.Now()
	if err := s.db.UpsertMatch(ctx, *match, s.clock.Now()); err != nil {
		logger.Error().Err(err).Msg("db upsert match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to update match")
	}

	logger.Info().Msg("match started")
	return httpx.Write(w, http.StatusOK, nil)
}

// FinishMatch godoc
// @Summary      Termine un match (passage à la saisie des scores)
// @Description  Passe un match de l’état "En cours" à "Manque Score" afin de permettre la saisie/validation des scores.
// @Tags         match
// @Produce      json
// @Param        id    path      string  true  "ID du match"
// @Success      200
// @Failure      400   {object}  models.Error  "ID manquant, mauvais état, ou utilisateur non inscrit au match"
// @Failure      401   {object}  models.Error  "Utilisateur non autorisé"
// @Failure      404   {object}  models.Error  "Match non trouvé"
// @Failure      500   {object}  models.Error  "Erreur serveur ou base de données"
// @Router       /match/{id}/finish [patch]
func (s *Service) FinishMatch(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "FinishMatch").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()
	id := chi.URLParam(r, "id")
	logger := baseLogger.With().Str("match_id", id).Logger()

	if id == "" {
		logger.Warn().Msg("missing match ID")
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	match, err := s.db.GetMatchById(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("db get match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if match == nil {
		logger.Warn().Msg("match not found")
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}
	if match.CurrentState != models.EnCours {
		logger.Warn().Str("state", string(match.CurrentState)).Msg("match not in EnCours")
		return httpx.WriteError(w, http.StatusBadRequest, "match is not in the right state")
	}
	if match.CreatorID != ai.UserID {
		msg := "user is not the match creator"
		logger.Warn().Str("creator_id", match.CreatorID).Msg(msg)
		return httpx.WriteError(w, http.StatusForbidden, msg)
	}

	userInMatch, err := s.db.IsUserInMatch(ctx, ai.UserID, id)
	if err != nil {
		logger.Error().Err(err).Msg("db check user in match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if !userInMatch {
		logger.Warn().Msg("user not in match")
		return httpx.WriteError(w, http.StatusBadRequest, "user is not in the match")
	}

	match.CurrentState = models.ManqueScore
	match.UpdatedAt = s.clock.Now()
	if err := s.db.UpsertMatch(ctx, *match, s.clock.Now()); err != nil {
		logger.Error().Err(err).Msg("db upsert match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to update match")
	}

	logger.Info().Msg("match finished (waiting scores)")
	return httpx.Write(w, http.StatusOK, nil)
}

// GetMatchVoteStatus godoc
// @Summary      Statut de vote des scores (par équipe) pour un match terminé
// @Description  Renvoie l'équipe du joueur, si son équipe a voté et le score voté (nullable), et la même info pour l'équipe adverse. Ne fonctionne que si le match est en statut "Termine".
// @Tags         match
// @Produce      json
// @Param        id   path      string  true  "ID du match"
// @Success      200  {object}  models.MatchVoteStatusResponse
// @Failure      400  {object}  models.Error  "ID manquant, mauvais état"
// @Failure      401  {object}  models.Error  "Utilisateur non autorisé"
// @Failure      404  {object}  models.Error  "Match ou utilisateur non trouvé"
// @Failure      500  {object}  models.Error  "Erreur serveur"
// @Router       /match/{id}/vote-status [get]
func (s *Service) GetMatchVoteStatus(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "GetMatchVoteStatus").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()
	matchID := chi.URLParam(r, "id")
	logger := baseLogger.With().Str("match_id", matchID).Logger()

	if matchID == "" {
		logger.Warn().Msg("missing match ID")
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	match, err := s.db.GetMatchById(ctx, matchID)
	if err != nil {
		logger.Error().Err(err).Msg("db get match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if match == nil {
		logger.Warn().Msg("match not found")
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}

	um, err := s.db.GetUserInMatch(ctx, ai.UserID, matchID)
	if err != nil {
		logger.Error().Err(err).Msg("db get user in match failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if um == nil {
		logger.Warn().Msg("user not in match")
		return httpx.WriteError(w, http.StatusNotFound, "user in this match not found")
	}
	myTeam := um.Team
	opTeam := 1
	if myTeam == 1 {
		opTeam = 2
	} else {
		opTeam = 1
	}

	myVote, err := s.db.GetScoreVoteByMatchAndTeam(ctx, matchID, myTeam)
	if err != nil {
		logger.Error().Err(err).Int("team", myTeam).Msg("db get vote for my team failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch team vote")
	}
	opVote, err := s.db.GetScoreVoteByMatchAndTeam(ctx, matchID, opTeam)
	if err != nil {
		logger.Error().Err(err).Int("team", opTeam).Msg("db get vote for opponent team failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch opponent team vote")
	}

	toStatus := func(v *models.DBMatchScoreVote) models.TeamVoteStatus {
		if v == nil {
			return models.TeamVoteStatus{HasVoted: false, Score: nil}
		}
		sp := &models.ScorePair{Score1: v.Score1, Score2: v.Score2}
		return models.TeamVoteStatus{HasVoted: true, Score: sp}
	}

	resp := models.MatchVoteStatusResponse{
		MatchID:    matchID,
		PlayerTeam: myTeam,
		MyTeam:     toStatus(myVote),
		Opponent:   toStatus(opVote),
	}

	logger.Info().Msg("vote status fetched")
	return httpx.Write(w, http.StatusOK, resp)
}
