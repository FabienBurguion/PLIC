package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"context"
	"encoding/json"
	"net/http"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func (s *Service) buildUserResponse(ctx context.Context, user *models.DBUsers, profilePictureUrl string) models.UserResponse {
	var (
		visitedFields int
		matchCount    int
		favSport      *models.Sport
		favField      *string
		sports        []models.Sport
		fields        []models.Field
		winrate       *int
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
	if wr, err := s.db.GetUserWinrate(ctx, user.Id); err == nil {
		winrate = wr
	}

	return models.UserResponse{
		Username:       user.Username,
		Bio:            user.Bio,
		CreatedAt:      user.CreatedAt,
		ProfilePicture: ptr(profilePictureUrl),
		CurrentFieldId: user.CurrentFieldId,
		VisitedFields:  visitedFields,
		NbMatches:      matchCount,
		Winrate:        winrate,
		FavoriteCity:   nil,
		FavoriteSport:  favSport,
		FavoriteField:  favField,
		Sports:         sports,
		Fields:         fields,
	}
}

func (s *Service) buildUserResponseFast(user *models.DBUsers, profilePictureURL string, st *models.UserStats) models.UserResponse {
	if st == nil {
		st = &models.UserStats{}
	}
	return models.UserResponse{
		Username:       user.Username,
		Bio:            user.Bio,
		CreatedAt:      user.CreatedAt,
		ProfilePicture: ptr(profilePictureURL),
		CurrentFieldId: user.CurrentFieldId,
		VisitedFields:  st.VisitedFields,
		NbMatches:      st.MatchCount,
		Winrate:        st.Winrate,
		FavoriteCity:   nil,
		FavoriteSport:  st.FavoriteSport,
		FavoriteField:  st.FavoriteField,
		Sports:         st.Sports,
		Fields:         st.Fields,
	}
}

// GetUserById godoc
// @Summary      Get a user by ID
// @Description  Retrieve user information, including profile picture and preferences
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
func (s *Service) GetUserById(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	logger := log.With().
		Str("method", "GetUserById").
		Str("user_id", ai.UserID).
		Str("path", r.URL.Path).
		Logger()

	if !ai.IsConnected {
		logger.Warn().Msg("unauthorized")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		logger.Warn().Msg("missing id in url params")
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	ctx := r.Context()
	user, err := s.db.GetUserById(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get user by id")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}
	if user == nil {
		logger.Info().Str("target_id", id).Msg("user not found")
		return httpx.WriteError(w, http.StatusNotFound, "user not found")
	}

	s3Resp, err := s.s3Service.GetProfilePicture(ctx, id)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to get profile picture from S3")
		s3Resp = &v4.PresignedHTTPRequest{URL: ""}
	}

	response := s.buildUserResponse(ctx, user, s3Resp.URL)
	logger.Info().Str("target_id", id).Msg("user fetched successfully")

	return httpx.Write(w, http.StatusOK, response)
}

// PatchUser godoc
// @Summary      Patch a user by ID
// @Description  Update user fields
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
	logger := log.With().
		Str("method", "PatchUser").
		Str("user_id", ai.UserID).
		Str("path", r.URL.Path).
		Logger()

	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if id == "" {
		logger.Warn().Msg("missing id in url params")
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	if !ai.IsConnected || ai.UserID != id {
		logger.Warn().Msg("unauthorized user tried to patch another account")
		return httpx.WriteError(w, http.StatusForbidden, "not authorized")
	}

	var req models.UserPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn().Err(err).Msg("invalid JSON in patch request")
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}

	if err := s.db.UpdateUser(ctx, req, id, s.clock.Now()); err != nil {
		logger.Error().Err(err).Msg("failed to update user in db")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}

	logger.Info().Str("target_id", id).Msg("user updated successfully")
	return httpx.Write(w, http.StatusOK, nil)
}

// DeleteUser godoc
// @Summary      Delete a user by ID
// @Description  Remove a user permanently
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
	logger := log.With().
		Str("method", "DeleteUser").
		Str("user_id", ai.UserID).
		Str("path", r.URL.Path).
		Logger()

	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if id == "" {
		logger.Warn().Msg("missing id in url params")
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	if !ai.IsConnected || ai.UserID != id {
		logger.Warn().Msg("unauthorized user tried to delete another account")
		return httpx.WriteError(w, http.StatusForbidden, "not authorized")
	}

	if err := s.db.DeleteUser(ctx, id); err != nil {
		logger.Error().Err(err).Msg("failed to delete user from db")
		return httpx.WriteError(w, http.StatusInternalServerError, "database error")
	}

	logger.Info().Str("target_id", id).Msg("user deleted successfully")
	return httpx.Write(w, http.StatusOK, nil)
}
