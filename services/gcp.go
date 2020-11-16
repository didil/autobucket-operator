package services

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/storage"
)

// GCPSvc GCP Service interface
type GCPSvc interface {
	CreateBucket(ctx context.Context, name string) error
}

// GCPService GCP Service struct
type GCPService struct {
	storageClient *storage.Client
}

// NewGCPService inits gcp service
func NewGCPService() (*GCPService, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("init gcp storage client: %v", err)
	}

	svc := &GCPService{
		storageClient: client,
	}

	return svc, nil
}

// CreateBucket creates a gcp bucket
func (svc *GCPService) CreateBucket(ctx context.Context, name string) error {
	cl := svc.storageClient

	bucket := cl.Bucket(name)

	_, err := bucket.Attrs(ctx)
	if err == nil {
		return nil // bucket already exists, noop
	}
	if err != nil && err != storage.ErrBucketNotExist {
		return fmt.Errorf("bucket attrs: %v", err)
	}

	err = bucket.Create(ctx, os.Getenv("GCP_PROJECT"), nil)
	if err != nil {
		return fmt.Errorf("create: %v", err)
	}

	return nil
}
