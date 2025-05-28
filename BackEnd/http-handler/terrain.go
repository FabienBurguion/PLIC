package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"
)

// GetAllTerrains godoc
// @Summary      Liste tous les terrains
// @Description  Retourne la liste de tous les terrains enregistrés en base de données
// @Tags         terrain
// @Produce      json
// @Success      200  {array}   models.DBCourt   "Liste des terrains"
// @Failure      500  {object}  models.Error       "Erreur lors de la récupération des terrains"
// @Router       /court/all [get]
func (s *Service) GetAllTerrains(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	terrains, err := s.db.GetAllTerrains(r.Context())
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch terrains: "+err.Error())
	}
	return httpx.Write(w, http.StatusOK, terrains)
}
