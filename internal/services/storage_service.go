package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// StorageService handles file storage operations using MinIO/S3
type StorageService struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

// NewStorageService creates a new storage service
func NewStorageService(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*StorageService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create MinIO client: %w", err)
	}

	return &StorageService{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		useSSL:     useSSL,
	}, nil
}

// EnsureBucketExists creates the bucket if it doesn't exist
func (s *StorageService) EnsureBucketExists(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("unable to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("unable to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile uploads a file to MinIO and returns the storage path and URL
func (s *StorageService) UploadFile(ctx context.Context, tenantName string, file *multipart.FileHeader) (string, string, error) {
	// Generate unique filename
	timestamp := time.Now().Format("2006/01/02")
	fileID := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	objectKey := fmt.Sprintf("%s/%s/%s%s", tenantName, timestamp, fileID, ext)

	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("unable to open file: %w", err)
	}
	defer src.Close()

	// Upload to MinIO
	_, err = s.client.PutObject(ctx, s.bucketName, objectKey, src, file.Size, minio.PutObjectOptions{
		ContentType: "application/pdf",
	})
	if err != nil {
		return "", "", fmt.Errorf("unable to upload file: %w", err)
	}

	// Generate URL
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucketName, objectKey)

	return objectKey, url, nil
}

