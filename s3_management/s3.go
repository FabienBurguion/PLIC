package s3_management

import (
	"bytes"
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type RealS3Service struct {
	Client *s3.Client
}

type S3Service interface {
	GetProfilePicture(ctx context.Context, userId string) (*v4.PresignedHTTPRequest, error)
	GetObject(ctx context.Context, bucketName string, objectKey string) (*v4.PresignedHTTPRequest, error)
	PutObject(ctx context.Context, bucketName string, objectKey string, buf *bytes.Buffer) error
}

func (s *RealS3Service) PutObject(ctx context.Context, bucketName string, objectKey string, buf *bytes.Buffer) error {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(http.DetectContentType(buf.Bytes())),
	})
	return err
}

func (s *RealS3Service) GetObject(ctx context.Context, bucketName string, objectKey string) (*v4.PresignedHTTPRequest, error) {
	log.Printf("Getting object %s", objectKey)
	presigner := s3.NewPresignClient(s.Client)
	return presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
}

func (s *RealS3Service) GetProfilePicture(ctx context.Context, userId string) (*v4.PresignedHTTPRequest, error) {
	return s.GetObject(ctx, "user-profil-pictures", userId+".png")
}
