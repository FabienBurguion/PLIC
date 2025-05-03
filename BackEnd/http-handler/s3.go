package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"PLIC/s3_management"
	"bytes"
	"errors"
	"github.com/aws/smithy-go"
	"log"
	"net/http"
)

// UploadImageToS3 godoc
// @Summary      Upload an image to S3
// @Description  Uploads an image file to an S3 bucket
// @Tags         upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        image formData file true "Image file to upload"
// @Success      201
// @Failure      400 {object} models.Error "Bad request or file not found"
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /image [post]
func (s *Service) UploadImageToS3(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	ctx := r.Context()
	bucketName := "test-plic"
	objectKey := bucketName + "file.png"

	file, _, err := r.FormFile("image")
	if err != nil {
		log.Printf("fichier non trouvé: %w", err)
		return httpx.WriteError(w, http.StatusBadRequest, httpx.BadRequestError)
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		log.Printf("lecture fichier: %v", err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	err = s3_management.PutObject(ctx, s.s3Client, bucketName, objectKey, buf)
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

// GetS3Image godoc
// @Summary      Get image URL from S3
// @Description  Retrieves a pre-signed URL to access an image stored in S3
// @Tags         upload
// @Produce      json
// @Success      201 {object} models.ImageUrl
// @Failure      500 {object} models.Error "Internal server error"
// @Router       /image [get]
func (s *Service) GetS3Image(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	ctx := r.Context()
	bucketName := "test-plic"

	objectKey := bucketName + "file.png"
	resp, err := s3_management.GetObject(ctx, s.s3Client, bucketName, objectKey)

	if err != nil {
		log.Println(err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	return httpx.Write(w, http.StatusCreated, models.ImageUrl{
		Url: resp.URL,
	})
}
