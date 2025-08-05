package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"context"
	"encoding/json"
	"log"
	"net/http"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/go-chi/chi/v5"
)

func (s *Service) BuildUserResponse(ctx context.Context, user *models.DBUsers, profilePictureUrl string) models.UserResponse {
	var (
		visitedFields int
		matchCount    int
		favSport      *models.Sport
		favField      *string
		sports        []models.Sport
		fields        []models.Field
	)

	if n, err := s.db.GetMatchCountByUserID(ctx, user.Id); err == nil {
		matchCount = n
	}
	if n, err := s.db.GetVisitedFieldCountByUserID(ctx, user.Id); err == nil {
		visitedFields = n
	}
	if c, err := s.db.GetFavoriteFieldByUserID(ctx, user.Id); err == nil {
		favField = c
	}
	if spt, err := s.db.GetFavoriteSportByUserID(ctx, user.Id); err == nil {
		favSport = spt
	}
	if lst, err := s.db.GetPlayedSportsByUserID(ctx, user.Id); err == nil {
		sports = lst
	}
	if lst, err := s.db.GetRankedFieldsByUserID(ctx, user.Id); err == nil {
		fields = lst
	}

	return models.UserResponse{
		Username:       user.Username,
		Bio:            user.Bio,
		CreatedAt:      user.CreatedAt,
		ProfilePicture: ptr(profilePictureUrl),
		CurrentFieldId: user.CurrentFieldId,
		VisitedFields:  visitedFields,
		NbMatches:      matchCount,
		Winrate:        ptr(80), // TODO
		FavoriteCity:   nil,
		FavoriteSport:  favSport,
		FavoriteField:  favField,
		Sports:         sports,
		Fields:         fields,
	}
}

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

	// --- BUILD FULL RESPONSE ---
	response := s.BuildUserResponse(ctx, user, s3Resp.URL)

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
