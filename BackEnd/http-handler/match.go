package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	"net/http"
)

func (s *Service) GetMatchByID(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	id := r.URL.Query().Get("id")
	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in query params")
	}

	match, err := s.db.GetMatchById(r.Context(), id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}
	if match == nil {
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}
	response := models.MatchResponse(*match)
	return httpx.Write(w, http.StatusOK, response)
}

func (s *Service) GetAllMatches(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	matches, err := s.db.GetAllMatches(r.Context())
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch matches: "+err.Error())
	}

	return httpx.Write(w, http.StatusOK, matches)
}

func (s *Service) CreateMatch(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	var match models.DBMatches

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&match); err != nil {
		return httpx.WriteError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
	}

	if match.Id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	if err := s.db.CreateMatch(r.Context(), match); err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to create match: "+err.Error())
	}

	return httpx.Write(w, http.StatusCreated, map[string]string{"status": "match created", "id": match.Id})
}
