package s3_management

import (
	"bytes"
	"context"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

type MockS3Service struct {
}

func (m *MockS3Service) GetProfilePicture(_ context.Context, userId string) (*v4.PresignedHTTPRequest, error) {
	return &v4.PresignedHTTPRequest{
		URL: userId + ".png",
	}, nil
}

func (m *MockS3Service) PutObject(_ context.Context, _ string, _ string, _ *bytes.Buffer) error {
	return nil
}

func (m *MockS3Service) GetObject(_ context.Context, _ string, objectKey string) (*v4.PresignedHTTPRequest, error) {
	return &v4.PresignedHTTPRequest{
		URL: objectKey + ".png",
	}, nil
}
