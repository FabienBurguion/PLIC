package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"net/http"
)

func (s *Service) GetAllTerrains(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	terrains, err := s.db.GetAllTerrains(r.Context())
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch terrains: "+err.Error())
	}
	return httpx.Write(w, http.StatusOK, terrains)
}
