package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"
)

// GetRankingByCourtId godoc
// @Summary      Classement ELO par terrain
// @Description  Retourne la liste des utilisateurs et leur ELO pour un court donné, triée par ELO croissant
// @Tags         ranking
// @Produce      json
// @Param        id   path      string  true  "Identifiant du court"
// @Success      200  {array}   models.CourtRankingResponse
// @Failure      400  {object}  models.Error  "ID manquant"
// @Failure      401  {object}  models.Error  "Utilisateur non autorisé"
// @Failure      500  {object}  models.Error  "Erreur serveur / base"
// @Router       /ranking/court/{id} [get]
func (s *Service) GetRankingByCourtId(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing court ID")
	}

	rows, err := s.db.GetRankingsByCourtID(ctx, id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch rankings: "+err.Error())
	}

	res := lo.Map(rows, func(rnk models.DBRanking, _ int) models.CourtRankingResponse {
		return models.CourtRankingResponse{
			UserID: rnk.UserID,
			Elo:    rnk.Elo,
		}
	})

	return httpx.Write(w, http.StatusOK, res)
}

// GetUserFields godoc
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
func (s *Service) GetUserFields(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	userID := chi.URLParam(r, "userId")
	if userID == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing userId in url params")
	}

	ctx := r.Context()
	fields, err := s.db.GetRankedFieldsByUserID(ctx, userID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch fields: "+err.Error())
	}

	return httpx.Write(w, http.StatusOK, fields)
}
