package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"sync"
)

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
func (s *Service) GetMatchByID(w http.ResponseWriter, r *http.Request, auth models.AuthInfo) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	if !auth.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	var (
		match    *models.DBMatches
		users    []models.DBUsers
		matchErr error
		usersErr error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	ctx := r.Context()

	go func() {
		defer wg.Done()
		match, matchErr = s.db.GetMatchById(ctx, id)
	}()

	go func() {
		defer wg.Done()
		users, usersErr = s.db.GetUsersByMatchId(ctx, id)
	}()

	wg.Wait()

	if matchErr != nil {
		log.Println("errored getting param by match:", matchErr)
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+matchErr.Error())
	}
	if match == nil {
		return httpx.WriteError(w, http.StatusNotFound, "param not found")
	}
	if usersErr != nil {
		log.Println("error getting users fom match", usersErr)
	}

	var profilePictures []string

	for _, user := range users {
		profilePicture, err := s.s3Service.GetProfilePicture(ctx, user.Id)
		if err != nil {
			log.Println("error getting profile picture:", err)
			profilePictures = append(profilePictures, "")
		} else {
			profilePictures = append(profilePictures, profilePicture.URL)
		}
	}

	response := match.ToMatchResponse(users, profilePictures)
	return httpx.Write(w, http.StatusOK, response)
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
func (s *Service) GetAllMatches(w http.ResponseWriter, r *http.Request, auth models.AuthInfo) error {
	ctx := r.Context()

	/*
		if !auth.IsConnected {
			return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
		}
	*/

	matches, err := s.db.GetAllMatches(ctx)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch matches: "+err.Error())
	}

	res := make([]models.MatchResponse, len(matches))

	var wg sync.WaitGroup
	wg.Add(len(matches))

	var mu sync.Mutex

	for i, match := range matches {
		go func(i int, match models.DBMatches) {
			defer wg.Done()

			users, userErr := s.db.GetUsersByMatchId(ctx, match.Id)
			if userErr != nil {
				log.Printf("warning: could not fetch users for match %s: %v", match.Id, userErr)
			}

			var profilePictures []string

			for _, user := range users {
				profilePicture, err := s.s3Service.GetProfilePicture(ctx, user.Id)
				if err != nil {
					log.Println("error getting profile picture:", err)
					profilePictures = append(profilePictures, "")
				} else {
					profilePictures = append(profilePictures, profilePicture.URL)
				}
			}

			mr := match.ToMatchResponse(users, profilePictures)

			mu.Lock()
			res[i] = mr
			mu.Unlock()
		}(i, match)
	}

	wg.Wait()

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
func (s *Service) CreateMatch(w http.ResponseWriter, r *http.Request, auth models.AuthInfo) error {
	var match models.MatchRequest

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&match); err != nil {
		return httpx.WriteError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
	}

	ctx := r.Context()

	if !auth.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	matchDb := match.ToDBMatches()

	if err := s.db.CreateMatch(r.Context(), matchDb); err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to create match: "+err.Error())
	}

	if err := s.db.AddUserToMatch(ctx, models.DBUserMatch{
		UserID:    auth.UserID,
		MatchID:   matchDb.Id,
		CreatedAt: s.clock.Now(),
	}); err != nil {
		log.Println("User creating match:", auth.UserID)
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to associate user to match: "+err.Error())
	}
	users, err := s.db.GetUsersByMatchId(ctx, matchDb.Id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch users: "+err.Error())
	}
	var profilePictures []string

	for _, user := range users {
		profilePicture, err := s.s3Service.GetProfilePicture(ctx, user.Id)
		if err != nil {
			log.Println("error getting profile picture:", err)
			profilePictures = append(profilePictures, "")
		} else {
			profilePictures = append(profilePictures, profilePicture.URL)
		}
	}
	response := matchDb.ToMatchResponse(users, profilePictures)

	return httpx.Write(w, http.StatusCreated, response)
}

// JoinMatch godoc
// @Summary      Un utilisateur rejoint un match
// @Description  Permet à un utilisateur authentifié de rejoindre un match existant, si ce n’est pas déjà fait
// @Tags         match
// @Produce      json
// @Param        id    path      string             true  "Identifiant du match"
// @Success      200   {object}  map[string]string  "Utilisateur a rejoint le match avec succès"
// @Failure      400   {object}  models.Error       "Identifiant manquant"
// @Failure      401   {object}  models.Error       "Utilisateur non autorisé"
// @Failure      404   {object}  models.Error       "Match non trouvé"
// @Failure      409   {object}  models.Error       "Utilisateur déjà inscrit au match"
// @Failure      500   {object}  models.Error       "Erreur lors de l'inscription de l'utilisateur au match"
// @Router       /match/join/{id} [post]
func (s *Service) JoinMatch(w http.ResponseWriter, r *http.Request, auth models.AuthInfo) error {
	ctx := r.Context()

	if !auth.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	matchID := chi.URLParam(r, "id")
	if matchID == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	match, err := s.db.GetMatchById(ctx, matchID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch match: "+err.Error())
	}
	if match == nil {
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}

	exists, err := s.db.IsUserInMatch(ctx, auth.UserID, matchID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check user in match: "+err.Error())
	}
	if exists {
		return httpx.WriteError(w, http.StatusConflict, "user already joined the match")
	}

	if err := s.db.AddUserToMatch(ctx, models.DBUserMatch{
		UserID:    auth.UserID,
		MatchID:   matchID,
		CreatedAt: s.clock.Now(),
	}); err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to join match: "+err.Error())
	}

	if err := s.db.IncrementMatchParticipants(ctx, matchID); err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to update participant count: "+err.Error())
	}

	return httpx.Write(w, http.StatusOK, map[string]string{
		"status": "joined match",
		"id":     matchID,
	})
}

// DeleteMatch godoc
// @Summary      Supprime un match
// @Description  Supprime un match via son ID
// @Tags         match
// @Produce      json
// @Param        id   path      string  true  "Identifiant du match à supprimer"
// @Success      200  {object}  map[string]string "Match supprimé"
// @Failure      400  {object}  models.Error      "ID manquant"
// @Failure      401  {object}  models.Error      "Utilisateur non autorisé"
// @Failure      500  {object}  models.Error      "Erreur lors de la suppression du match"
// @Router       /match/{id} [delete]
func (s *Service) DeleteMatch(w http.ResponseWriter, r *http.Request, auth models.AuthInfo) error {
	if !auth.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	matchID := chi.URLParam(r, "id")
	if matchID == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	ctx := r.Context()
	if err := s.db.DeleteMatch(ctx, matchID); err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to delete match: "+err.Error())
	}

	return httpx.Write(w, http.StatusOK, map[string]string{
		"status": "deleted match",
		"id":     matchID,
	})
}
