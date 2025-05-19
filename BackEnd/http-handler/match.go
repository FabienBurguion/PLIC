package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	"net/http"
)

// GetMatchByID godoc
// @Summary      Récupère un match par son ID
// @Description  Retourne les informations d’un match en fonction de son identifiant passé en paramètre de requête
// @Tags         match
// @Produce      json
// @Param        id   query     string  true  "Identifiant du match"
// @Success      200  {object}  models.MatchResponse "Match trouvé"
// @Failure      400  {object}  models.Error         "ID manquant ou invalide"
// @Failure      404  {object}  models.Error         "Match non trouvé"
// @Failure      500  {object}  models.Error         "Erreur serveur ou base de données"
// @Router       /match [get]
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

// GetAllMatches godoc
// @Summary      Liste tous les matchs
// @Description  Retourne la liste complète de tous les matchs stockés en base
// @Tags         match
// @Produce      json
// @Success      200  {array}   models.MatchResponse "Liste des matchs"
// @Failure      500  {object}  models.Error          "Erreur serveur lors de la récupération des matchs"
// @Router       /match/all [get]
func (s *Service) GetAllMatches(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	matches, err := s.db.GetAllMatches(r.Context())
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch matches: "+err.Error())
	}

	res := make([]models.MatchResponse, len(matches))

	for i, match := range matches {
		res[i] = match.ToMatchResponse()
	}

	return httpx.Write(w, http.StatusOK, res)
}

// CreateMatch godoc
// @Summary      Crée un nouveau match
// @Description  Enregistre un nouveau match en base de données à partir des données fournies en JSON
// @Tags         match
// @Accept       json
// @Produce      json
// @Param        match  body      models.DBMatches  true  "Objet match à créer"
// @Success      201    {object}  map[string]string "Match créé avec succès"
// @Failure      400    {object}  models.Error      "Données invalides ou champ ID manquant"
// @Failure      500    {object}  models.Error      "Erreur lors de la création du match"
// @Router       /match [post]
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
