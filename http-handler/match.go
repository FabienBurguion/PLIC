package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Service) buildMatchesResponse(ctx context.Context, matches []models.DBMatches) []models.MatchResponse {
	responses := make([]models.MatchResponse, 0, len(matches))

	for _, match := range matches {
		users, userErr := s.db.GetUsersByMatchId(ctx, match.Id)
		if userErr != nil {
			log.Printf("warning: could not fetch users for match %s: %v", match.Id, userErr)
			continue
		}

		profilePictures := make([]string, len(users))
		var wg sync.WaitGroup

		for i, user := range users {
			wg.Add(1)
			go func(i int, user models.DBUsers) {
				defer wg.Done()
				pic, err := s.s3Service.GetProfilePicture(ctx, user.Id)
				if err != nil {
					log.Printf("error getting profile picture for user %s: %v", user.Id, err)
					profilePictures[i] = ""
				} else {
					profilePictures[i] = pic.URL
				}
			}(i, user)
		}
		wg.Wait()

		buildErr, matchResponse := s.buildMatchResponse(ctx, match, users, profilePictures)
		if buildErr != nil {
			log.Printf("warning: could not build match response for match %s: %v", match.Id, buildErr)
			continue
		}

		responses = append(responses, matchResponse)
	}

	return responses
}

func (s *Service) buildMatchResponse(ctx context.Context, match models.DBMatches, users []models.DBUsers, profilePictures []string) (error, models.MatchResponse) {
	userResponses := make([]models.UserResponse, len(users))
	for i, u := range users {
		var profilePicture string
		if i >= len(profilePictures) {
			log.Printf("profilePictures index out of range: i=%d, len=%d", i, len(profilePictures))
			profilePicture = ""
		} else {
			profilePicture = profilePictures[i]
		}
		userResponses[i] = s.BuildUserResponse(ctx, &u, profilePicture)
	}

	court, err := s.db.GetCourtByID(ctx, match.CourtID)
	if err != nil {
		log.Println("error getting court:", err)
		return fmt.Errorf("error getting court %s: %s", match.CourtID, err.Error()), models.MatchResponse{}
	}

	if court == nil {
		log.Println("error: no court found for match", match.Id)
		return fmt.Errorf("error: no court found for match %s", match.Id), models.MatchResponse{}
	}

	return nil, models.MatchResponse{
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
	}
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
	id := chi.URLParam(r, "id")
	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	if !ai.IsConnected {
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
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}
	if usersErr != nil {
		log.Println("error getting users fom match", usersErr)
	}

	var (
		profilePictures = make([]string, len(users))
		wg2             sync.WaitGroup
	)

	for i, user := range users {
		wg2.Add(1)
		go func(i int, user models.DBUsers) {
			defer wg2.Done()
			profilePicture, err := s.s3Service.GetProfilePicture(ctx, user.Id)
			if err != nil {
				log.Println("error getting profile picture:", err)
				profilePictures[i] = ""
			} else {
				profilePictures[i] = profilePicture.URL
			}
		}(i, user)
	}

	wg2.Wait()

	err, response := s.buildMatchResponse(ctx, *match, users, profilePictures)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to build match response: "+err.Error())
	}
	return httpx.Write(w, http.StatusOK, response)
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
	userId := chi.URLParam(r, "userId")
	if userId == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing userId in url params")
	}

	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()

	dbMatches, err := s.db.GetMatchesByUserID(ctx, userId)
	if err != nil {
		log.Println("error getting matches:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}

	res := s.buildMatchesResponse(ctx, dbMatches)

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
// @Router       /match/court/{courtId} [get]
func (s *Service) GetMatchesByCourtId(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	courtID := chi.URLParam(r, "courtId")
	if courtID == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing courtId in url params")
	}

	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()

	matches, err := s.db.GetMatchesByCourtId(ctx, courtID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}

	res := s.buildMatchesResponse(ctx, matches)

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
	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()

	matches, err := s.db.GetAllMatches(ctx)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch matches: "+err.Error())
	}

	res := s.buildMatchesResponse(ctx, matches)

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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	if err := decoder.Decode(&match); err != nil {
		return httpx.WriteError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
	}

	if match.NbreParticipant < 2 || match.NbreParticipant%2 != 0 {
		return httpx.WriteError(w, http.StatusBadRequest, "invalid number of participant")
	}

	ctx := r.Context()

	if !auth.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	court, err := s.db.GetCourtByID(ctx, match.CourtID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch court: "+err.Error())
	}
	if court == nil {
		return httpx.WriteError(w, http.StatusBadRequest, "court not found")
	}

	matchDb := match.ToDBMatches(s.clock.Now())

	if err := s.db.CreateMatch(ctx, matchDb); err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to create match: "+err.Error())
	}

	existing, err := s.db.GetRankingByUserAndCourt(ctx, auth.UserID, match.CourtID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check ranking: "+err.Error())
	}
	if existing == nil {
		if err := s.db.InsertRanking(ctx, models.DBRanking{
			UserID:    auth.UserID,
			CourtID:   match.CourtID,
			Elo:       DefaultElo,
			CreatedAt: s.clock.Now(),
			UpdatedAt: s.clock.Now(),
		}); err != nil {
			return httpx.WriteError(w, http.StatusInternalServerError, "failed to create default ranking: "+err.Error())
		}
	}

	if err := s.db.CreateUserMatch(ctx, models.DBUserMatch{
		UserID:    auth.UserID,
		MatchID:   matchDb.Id,
		Team:      1, // When you create a match, you join the first team
		CreatedAt: s.clock.Now(),
	}); err != nil {
		log.Println("User creating match:", auth.UserID)
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to associate user to match: "+err.Error())
	}

	return httpx.Write(w, http.StatusCreated, models.CreateMatchResponse{
		Id: matchDb.Id,
	})
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

	var matchRequest models.JoinMatchRequest

	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	if err := decoder.Decode(&matchRequest); err != nil {
		return httpx.WriteError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
	}

	match, err := s.db.GetMatchById(ctx, matchID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch match: "+err.Error())
	}
	if match == nil {
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}

	if match.CurrentState != models.ManqueJoueur {
		return httpx.WriteError(w, http.StatusBadRequest, "match is not in the right state")
	}

	exists, err := s.db.IsUserInMatch(ctx, auth.UserID, matchID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check user in match: "+err.Error())
	}
	if exists {
		return httpx.WriteError(w, http.StatusConflict, "user already joined the match")
	}

	count, err := s.db.CountUsersByMatchAndTeam(ctx, matchID, matchRequest.Team)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to count users by match and team: "+err.Error())
	}
	if count >= match.ParticipantNber/2 {
		return httpx.WriteError(w, http.StatusBadRequest, "this team is full")
	}

	existing, err := s.db.GetRankingByUserAndCourt(ctx, auth.UserID, match.CourtID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check ranking: "+err.Error())
	}
	if existing == nil {
		if err := s.db.InsertRanking(ctx, models.DBRanking{
			UserID:    auth.UserID,
			CourtID:   match.CourtID,
			Elo:       DefaultElo,
			CreatedAt: s.clock.Now(),
			UpdatedAt: s.clock.Now(),
		}); err != nil {
			return httpx.WriteError(w, http.StatusInternalServerError, "failed to create default ranking: "+err.Error())
		}
	}

	if err := s.db.CreateUserMatch(ctx, models.DBUserMatch{
		UserID:    auth.UserID,
		MatchID:   matchID,
		Team:      matchRequest.Team,
		CreatedAt: s.clock.Now(),
	}); err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to join match: "+err.Error())
	}

	newCount, err := s.db.CountUsersByMatch(ctx, matchID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to count users by match: "+err.Error())
	}
	if newCount == match.ParticipantNber {
		match.CurrentState = models.Valide
	}

	match.UpdatedAt = s.clock.Now()
	err = s.db.UpsertMatch(ctx, *match)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to update match: "+err.Error())
	}

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

	return httpx.Write(w, http.StatusOK, nil)
}

func (s *Service) getOrCreateRanking(ctx context.Context, userID, courtID string, now time.Time) (models.DBRanking, error) {
	rk, err := s.db.GetRankingByUserAndCourt(ctx, userID, courtID)
	if err != nil {
		return models.DBRanking{}, err
	}
	if rk != nil {
		return *rk, nil
	}
	newRk := models.DBRanking{
		UserID:    userID,
		CourtID:   courtID,
		Elo:       DefaultElo,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.db.InsertRanking(ctx, newRk); err != nil {
		return models.DBRanking{}, err
	}
	return newRk, nil
}

func (s *Service) applyEloForMatch(ctx context.Context, match models.DBMatches, score1, score2 int) error {
	userMatches, err := s.db.GetUserMatchesByMatchID(ctx, match.Id)
	if err != nil {
		return err
	}
	if len(userMatches) == 0 {
		return nil
	}

	team1Users := make([]string, 0)
	team2Users := make([]string, 0)
	for _, um := range userMatches {
		if um.Team == 1 {
			team1Users = append(team1Users, um.UserID)
		} else if um.Team == 2 {
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
			rk, err := s.getOrCreateRanking(ctx, uid, match.CourtID, now)
			if err != nil {
				return nil, err
			}
			out = append(out, rk)
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
// @Router       /match/{id}/score [patch]
func (s *Service) UpdateMatchScore(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	var req models.UpdateScoreRequest
	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	if err := decoder.Decode(&req); err != nil {
		return httpx.WriteError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
	}

	match, err := s.db.GetMatchById(ctx, id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}
	if match == nil {
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}

	if match.CurrentState != models.ManqueScore {
		return httpx.WriteError(w, http.StatusBadRequest, "match is not in the right state")
	}

	userMatch, err := s.db.GetUserInMatch(ctx, ai.UserID, id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}
	if userMatch == nil {
		return httpx.WriteError(w, http.StatusNotFound, "user in this match not found")
	}

	hasSameTeamOtherVote, err := s.db.HasOtherTeamVote(ctx, id, userMatch.Team, ai.UserID)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check team vote: "+err.Error())
	}
	if hasSameTeamOtherVote {
		return httpx.WriteError(w, http.StatusBadRequest, "this team already has a vote")
	}

	if err := s.db.UpsertMatchScoreVote(ctx, models.DBMatchScoreVote{
		MatchId: id,
		UserId:  ai.UserID,
		Team:    userMatch.Team,
		Score1:  req.Score1,
		Score2:  req.Score2,
	}); err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to upsert score vote: "+err.Error())
	}

	hasConsensus, err := s.db.HasConsensusScore(ctx, id, userMatch.Team, req.Score1, req.Score2)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to check consensus: "+err.Error())
	}

	match.Score1 = &req.Score1
	match.Score2 = &req.Score2
	match.UpdatedAt = s.clock.Now()

	if hasConsensus {
		match.CurrentState = models.Termine

		if err := s.applyEloForMatch(ctx, *match, req.Score1, req.Score2); err != nil {
			return httpx.WriteError(w, http.StatusInternalServerError, "failed to update rankings: "+err.Error())
		}
	}

	if err := s.db.UpsertMatch(ctx, *match); err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to update match: "+err.Error())
	}

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
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	match, err := s.db.GetMatchById(ctx, id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}
	if match == nil {
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}
	if match.CurrentState != models.Valide {
		return httpx.WriteError(w, http.StatusBadRequest, "match is not in the right state")
	}

	userInMatch, err := s.db.IsUserInMatch(ctx, ai.UserID, id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}
	if !userInMatch {
		return httpx.WriteError(w, http.StatusBadRequest, "user is not in the match")
	}

	match.Date = s.clock.Now()
	match.CurrentState = models.EnCours
	match.UpdatedAt = s.clock.Now()
	err = s.db.UpsertMatch(ctx, *match)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to update match: "+err.Error())
	}

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
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing match ID")
	}

	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	match, err := s.db.GetMatchById(ctx, id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}
	if match == nil {
		return httpx.WriteError(w, http.StatusNotFound, "match not found")
	}
	if match.CurrentState != models.EnCours {
		return httpx.WriteError(w, http.StatusBadRequest, "match is not in the right state")
	}

	userInMatch, err := s.db.IsUserInMatch(ctx, ai.UserID, id)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}
	if !userInMatch {
		return httpx.WriteError(w, http.StatusBadRequest, "user is not in the match")
	}

	match.CurrentState = models.ManqueScore
	match.UpdatedAt = s.clock.Now()
	err = s.db.UpsertMatch(ctx, *match)
	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "failed to update match: "+err.Error())
	}

	return httpx.Write(w, http.StatusOK, nil)
}
