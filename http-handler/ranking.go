package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// GetRankingByCourtId godoc
// @Summary      Classement ELO par terrain
// @Description  Retourne la liste des utilisateurs et leur ELO pour un court donné, triée par ELO croissant
// @Tags         ranking
// @Produce      json
// @Param        id   path      string  true  "Identifiant du court"
// @Param        body body      models.CourtRankingRequest true  "Nouveaux scores"
// @Success      200  {array}   models.CourtRankingResponse
// @Failure      400  {object}  models.Error  "ID manquant"
// @Failure      401  {object}  models.Error  "Utilisateur non autorisé"
// @Failure      500  {object}  models.Error  "Erreur serveur / base"
// @Router       /ranking/court/{id} [get]
func (s *Service) GetRankingByCourtId(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "GetRankingByCourtId").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()
	id := chi.URLParam(r, "id")

	logger := baseLogger.With().Str("court_id", id).Logger()

	if id == "" {
		logger.Warn().Msg("missing court ID")
		return httpx.WriteError(w, http.StatusBadRequest, "missing court ID")
	}

	var req models.CourtRankingRequest
	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) { _ = Body.Close() }(r.Body)
	if err := decoder.Decode(&req); err != nil {
		logger.Warn().Err(err).Msg("invalid JSON body")
		return httpx.WriteError(w, http.StatusBadRequest, "invalid JSON")
	}

	rows, err := s.db.GetRankingsByCourtID(ctx, id, req.Sport)
	if err != nil {
		logger.Error().Err(err).Msg("db get rankings by court failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch rankings")
	}

	res := lo.Map(rows, func(rnk models.DBRanking, _ int) models.CourtRankingResponse {
		return models.CourtRankingResponse{
			UserID: rnk.UserID,
			Elo:    rnk.Elo,
		}
	})

	logger.Info().Int("count", len(res)).Msg("rankings fetched")
	return httpx.Write(w, http.StatusOK, res)
}

// GetRankedFieldsByUserID godoc
// @Summary      Liste des terrains (fields) d’un utilisateur
// @Description  Retourne uniquement la liste des fields associés à un utilisateur (ex: terrains classés/évalués)
// @Tags         user
// @Produce      json
// @Param        userId   path      string  true  "Identifiant de l'utilisateur"
// @Success      200  {array}   models.Field
// @Failure      400  {object}  models.Error  "userId manquant"
// @Failure      401  {object}  models.Error  "Utilisateur non autorisé"
// @Failure      500  {object}  models.Error  "Erreur serveur / base"
// @Router       /ranking/user/{userId} [get]
func (s *Service) GetRankedFieldsByUserID(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	baseLogger := log.With().
		Str("method", "GetRankedFieldsByUserID").
		Str("user_id", ai.UserID).
		Logger()

	if !ai.IsConnected {
		baseLogger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	userID := chi.URLParam(r, "userId")
	logger := baseLogger.With().Str("target_user_id", userID).Logger()

	if userID == "" {
		logger.Warn().Msg("missing userId in url params")
		return httpx.WriteError(w, http.StatusBadRequest, "missing userId in url params")
	}

	ctx := r.Context()
	fields, err := s.db.GetRankedFieldsByUserID(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Msg("db get ranked fields by user failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch fields")
	}

	logger.Info().Int("count", len(fields)).Msg("user fields fetched")
	return httpx.Write(w, http.StatusOK, fields)
}
