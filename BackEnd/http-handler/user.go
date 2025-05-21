package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"PLIC/s3_management"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

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
func (s *Service) GetUserById(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	user, err := s.db.GetUserById(r.Context(), id)
	if err != nil {
		log.Println("errored getting user by id:", err)
		return httpx.WriteError(w, http.StatusInternalServerError, "database error: "+err.Error())
	}

	if user == nil {
		return httpx.WriteError(w, http.StatusNotFound, "user not found")
	}

	s3Response, err := s3_management.GetProfilePicture(r.Context(), s.s3Client, id)
	if err != nil {
		log.Println("error getting profile picture:", err)
	}

	res := models.UserResponse{
		Username:       user.Username,
		ProfilePicture: ptr(s3Response.URL),
		Bio:            user.Bio,
		CreatedAt:      user.CreatedAt,
		VisitedFields:  0,
		Winrate:        100,
		FavoriteCity:   "a wonderful city",
		FavoriteSport:  "a wonderful sport",
		FavoriteField:  "a wonderful field",
		Sports: []models.Sport{
			models.Basket,
			models.Foot,
		},
		Fields: []models.Field{
			{
				Ranking: 1,
				Name:    "a wonderful field",
				Score:   1000,
			},
		},
	}

	return httpx.Write(w, http.StatusOK, res)
}
