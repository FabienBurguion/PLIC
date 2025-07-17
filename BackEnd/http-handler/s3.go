package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"bytes"
	"errors"
	"github.com/aws/smithy-go"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

// UploadProfilePictureToS3 godoc
// @Summary      Upload a profile picture to S3
// @Description  Uploads a profile picture to an S3 bucket
// @Tags         upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        image formData file true "Image file to upload"
// @Success      201
// @Failure      400 {object} models.Error "Bad request or file not found"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /profile_picture/{id} [post]
func (s *Service) UploadProfilePictureToS3(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	ctx := r.Context()
	bucketName := "param-profil-pictures"

	id := chi.URLParam(r, "id")
	if id == "" {
		return httpx.WriteError(w, http.StatusBadRequest, "missing id in url params")
	}

	if !ai.IsConnected {
		return httpx.WriteError(w, http.StatusForbidden, "not authorized")
	}

	if ai.UserID != id {
		return httpx.WriteError(w, http.StatusBadRequest, "bad request")
	}

	objectKey := id + ".png"

	file, _, err := r.FormFile("image")
	if err != nil {
		log.Printf("fichier non trouvé: %v", err)
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("⚠️ Erreur lors de la fermeture du fichier : %v", err)
		}
	}()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		log.Printf("lecture fichier: %v", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	err = s.s3Service.PutObject(ctx, bucketName, objectKey, buf)
	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("🛑 OperationError: service=%s, operation=%s, err=%v\n", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		log.Printf("🛑 Erreur détaillée upload S3: %#v\n", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	return httpx.Write(w, http.StatusCreated, nil)
}
