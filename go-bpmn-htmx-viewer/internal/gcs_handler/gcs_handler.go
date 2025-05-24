package gcs_handler

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
)

// FetchBPMNFromGCS fetches a file from Google Cloud Storage.
func FetchBPMNFromGCS(ctx context.Context, gcsURI string) (string, error) {
	if !strings.HasPrefix(gcsURI, "gs://") {
		return "", fmt.Errorf("invalid GCS URI: must start with gs://, got %s", gcsURI)
	}

	trimmedURI := strings.TrimPrefix(gcsURI, "gs://")
	parts := strings.SplitN(trimmedURI, "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid GCS URI format: expected gs://bucket-name/object-path, got %s", gcsURI)
	}
	bucketName := parts[0]
	objectName := parts[1]

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer client.Close()

	obj := client.Bucket(bucketName).Object(objectName)
	rc, err := obj.NewReader(ctx)
	if err != nil {
		// Consider checking for storage.ErrObjectNotExist specifically if desired
		return "", fmt.Errorf("failed to create GCS object reader for %s: %w", gcsURI, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("failed to read data from GCS object %s: %w", gcsURI, err)
	}

	return string(data), nil
}
