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

func (s *Service) UploadImageToS3(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	ctx := r.Context()
	bucketName := "test-plic"
	objectKey := bucketName + "file.png"

	file, _, err := r.FormFile("image")
	if err != nil {
		log.Printf("fichier non trouvé: %w", err)
		return httpx.WriteError(w, http.StatusNotFound, httpx.NotFoundError)
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		log.Printf("lecture fichier: %w", err)
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
