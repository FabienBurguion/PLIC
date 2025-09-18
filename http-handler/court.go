package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetAllCourts godoc
// @Summary      Liste tous les terrains
// @Description  Retourne la liste de tous les terrains enregistrés en base de données
// @Tags         terrain
// @Produce      json
// @Success      200  {array}   models.DBCourt   "Liste des terrains"
// @Failure      500  {object}  models.Error       "Erreur lors de la récupération des terrains"
// @Router       /court/all [get]
func (s *Service) GetAllCourts(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	if !ai.IsConnected{
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}
	terrains, err := s.db.GetAllCourts(r.Context())
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch terrains: "+err.Error())
	}
	return httpx.Write(w, http.StatusOK, terrains)
}

// GetCourtByID godoc
// @Summary      Récupère un terrain par son ID
// @Description  Retourne les informations d’un terrain (court) en fonction de son identifiant passé dans l’URL
// @Tags         terrain
// @Produce      json
// @Param        id   path      string  true  "Identifiant du terrain"
// @Success      200  {object}  models.DBCourt  "Terrain trouvé"
// @Failure      400  {object}  models.Error    "ID manquant ou invalide"
// @Failure      404  {object}  models.Error    "Terrain non trouvé"
// @Failure      500  {object}  models.Error    "Erreur serveur ou base de données"
// @Router       /court/{id} [get]
func (s *Service) GetCourtByID(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	if !ai.IsConnected{
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	court, err := s.db.GetCourtByID(r.Context(), id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch court: "+err.Error())
	}
	if court == nil {
		return httpx.WriteError(w, http.StatusNotFound, "court not found")
	}

	return httpx.Write(w, http.StatusOK, court)
}
