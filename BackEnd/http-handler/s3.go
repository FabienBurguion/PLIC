package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"bytes"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(http.DetectContentType(buf.Bytes())),
	})
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
	presigner := s3.NewPresignClient(s.s3Client)
	resp, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		log.Println(err)
		return httpx.WriteError(w, http.StatusInternalServerError, httpx.InternalServerError)
	}

	fmt.Println("🔗 URL S3 presignée :", resp.URL)

	testResp, err := http.Get(resp.URL)
	if err != nil {
		log.Println("❌ URL invalide :", err)
	} else {
		log.Println("✅ Code HTTP presigned:", testResp.StatusCode)
	}

	return httpx.Write(w, http.StatusCreated, models.ImageUrl{
		Url: resp.URL,
	})
}
