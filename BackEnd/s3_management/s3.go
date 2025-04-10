package s3_management

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/http"
)

func PutObject(ctx context.Context, s3Client *s3.Client, bucketName string, objectKey string, buf *bytes.Buffer) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(http.DetectContentType(buf.Bytes())),
	})
	return err
}

func GetObject(ctx context.Context, s3Client *s3.Client, bucketName string, objectKey string) (*v4.PresignedHTTPRequest, error) {
	presigner := s3.NewPresignClient(s3Client)
	return presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
}
