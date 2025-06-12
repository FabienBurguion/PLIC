package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"encoding/json"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"sync"
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

	var (
		user       *models.DBUsers
		s3Response *v4.PresignedHTTPRequest
		userErr    error
		s3Err      error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	ctx := r.Context()

	go func() {
		defer wg.Done()
		user, userErr = s.db.GetUserById(ctx, id)
	}()

	go func() {
		defer wg.Done()
		s3Response, s3Err = s.s3Service.GetProfilePicture(ctx, id)
	}()

	wg.Wait()

	if userErr != nil {
		log.Println("errored getting param by id:", userErr)
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+userErr.Error())
	}

	if user == nil {
		return httpx.WriteError(w, http.StatusNotFound, "param not found")
	}

	if s3Err != nil {
		log.Println("error getting profile picture:", s3Err)
	}

	res := models.UserResponse{
		Username:       user.Username,
		ProfilePicture: ptr(s3Response.URL),
		Bio:            user.Bio,
		CreatedAt:      user.CreatedAt,
		VisitedFields:  0,                        // TODO NO HARDCODE
		Winrate:        ptr(100),                 // TODO NO HARDCODE
		FavoriteCity:   ptr("a wonderful city"),  // TODO NO HARDCODE
		FavoriteSport:  ptr(models.Foot),         // TODO NO HARDCODE
		FavoriteField:  ptr("a wonderful field"), // TODO NO HARDCODE
		Sports: []models.Sport{ // TODO NO HARDCODE
			models.Basket,
			models.Foot,
		},
		Fields: []models.Field{ // TODO NO HARDCODE
			{
				Ranking: 1,
				Name:    "a wonderful field",
				Score:   1000,
			},
		},
	}

	return httpx.Write(w, http.StatusOK, res)
}

// PatchUser godoc
// @Summary      Patch a user by ID
// @Description  Patch a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
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
