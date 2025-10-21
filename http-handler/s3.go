package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"bytes"
	"errors"
	"net/http"

	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
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
// @Failure      401 {object} models.Error "Unauthorized"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /profile_picture/{id} [post]
func (s *Service) UploadProfilePictureToS3(w http.ResponseWriter, r *http.Request, ai models.AuthInfo) error {
	logger := log.With().
		Str("method", "UploadProfilePictureToS3").
		Str("user_id", ai.UserID).
		Str("path", r.URL.Path).
		Logger()

	if !ai.IsConnected {
		logger.Warn().Msg("unauthorized user tried to upload profile picture")
		return httpx.WriteError(w, http.StatusUnauthorized, "not authorized")
	}

	ctx := r.Context()
	bucketName := "user-profil-pictures"

	objectKey := ai.UserID + ".png"
	logger = logger.With().Str("bucket", bucketName).Str("object_key", objectKey).Logger()

	file, _, err := r.FormFile("image")
	if err != nil {
		msg := "Image file not found or incorrect format"
		logger.Warn().Err(err).Msg(msg)
		return httpx.WriteError(w, http.StatusBadRequest, msg)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Warn().Err(err).Msg("error closing file")
		}
	}()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		logger.Error().Err(err).Msg("failed to read uploaded file")
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	if err := s.s3Service.PutObject(ctx, bucketName, objectKey, buf); err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			logger.Error().
				Str("service", oe.Service()).
				Str("operation", oe.Operation()).
				Err(oe.Unwrap()).
				Msg("aws operation error")
		} else {
			logger.Error().Err(err).Msg("failed to upload to S3")
		}
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	logger.Info().Msg("profile picture uploaded successfully")
	return httpx.Write(w, http.StatusCreated, nil)
}
