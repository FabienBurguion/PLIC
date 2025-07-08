package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

// GetUserById godoc
// @Summary      Get a param by ID
// @Description  Retrieve param information, including profile picture and preferences
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} models.UserResponse
// @Failure      400 {object} models.Error "Missing ID in URL params"
// @Failure      404 {object} models.Error "User not found"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /users/{id} [get]
// @Security     BearerAuth
func (s *Service) GetUserById(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	ctx := r.Context()

	// --- USER ---
	user, err := s.db.GetUserById(ctx, id)
	if err != nil {
		log.Println("error getting user by id:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}
	if user == nil {
		return httpx.WriteError(w, http.StatusNotFound, "user not found")
	}

	// --- PROFILE PICTURE ---
	s3Resp, err := s.s3Service.GetProfilePicture(ctx, id)
	if err != nil {
		log.Println("error getting profile picture:", err)
		s3Resp = &v4.PresignedHTTPRequest{URL: ""}
	}

	// --- MATCH COUNT ---
	matchCount, err := s.db.GetMatchCountByUserID(ctx, id)
	if err != nil {
		log.Println("error getting match count:", err)
		matchCount = 0
	}

	// --- VISITED FIELDS ---
	visitedFields, err := s.db.GetVisitedFieldCountByUserID(ctx, id)
	if err != nil {
		log.Println("error getting visited fields:", err)
		visitedFields = 0
	}

	// --- FAVORITE FIELD ---
	favCity, err := s.db.GetFavoriteFieldByUserID(ctx, id)
	if err != nil {
		log.Println("error getting favorite city:", err)
		favCity = nil
	}

	// --- FAVORITE SPORT ---
	favSport, err := s.db.GetFavoriteSportByUserID(ctx, id)
	if err != nil {
		log.Println("error getting favorite sport:", err)
		favSport = nil
	}

	// --- SPORTS ---
	sports, err := s.db.GetPlayedSportsByUserID(ctx, id)
	if err != nil {
		log.Println("error getting played sports:", err)
		sports = []models.Sport{}
	}

	// --- FIELDS ---
	fields, err := s.db.GetRankedFieldsByUserID(ctx, id)
	if err != nil {
		log.Println("error getting ranked fields:", err)
		fields = []models.Field{}
	}

	// --- BUILD RESPONSE ---
	response := models.UserResponse{
		Username:       user.Username,
		ProfilePicture: ptr(s3Resp.URL),
		Bio:            user.Bio,
		CurrentFieldId: user.CurrentFieldId,
		CreatedAt:      user.CreatedAt,
		VisitedFields:  visitedFields,
		NbMatches:      matchCount,
		Winrate:        ptr(100), // TODO: calculer
		FavoriteCity:   nil,
		FavoriteSport:  favSport,
		FavoriteField:  favCity,
		Sports:         sports,
		Fields:         fields,
	}

	return httpx.Write(w, http.StatusOK, response)
}

// PatchUser godoc
// @Summary      Patch a user by ID
// @Description  Patch a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        body body models.UserPatchRequest true "User fields to update"
// @Success      200
// @Failure      400 {object} models.Error "Missing ID in URL params"
// @Failure      403 {object} models.Error "Incorrect rights"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /users/{id} [patch]
// @Security     BearerAuth
func (s *Service) PatchUser(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	if !ai.IsConnected || ai.UserID != id {
		return httpx.WriteError(w, http.StatusForbidden, "not authorized")
	}

	var req models.UserPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Erreur JSON:", err)
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}

	err := s.db.UpdateUser(ctx, req, id, s.clock.Now())

	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}

	return httpx.Write(w, http.StatusOK, nil)
}

// DeleteUser godoc
// @Summary      Delete a user by ID
// @Description  Delete a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200
// @Failure      400 {object} models.Error "Missing ID in URL params"
// @Failure      403 {object} models.Error "Incorrect rights"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /users/{id} [delete]
// @Security     BearerAuth
func (s *Service) DeleteUser(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	if !ai.IsConnected || ai.UserID != id {
		return httpx.WriteError(w, http.StatusForbidden, "not authorized")
	}

	err := s.db.DeleteUser(ctx, id)

	if err != nil {
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}

	return httpx.Write(w, http.StatusOK, nil)
}
