package gcs

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"

	"github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher/pkg/common"
)

func Fetch(ctx context.Context, source, destFileName string) error {
	location := "gs://" + source
	bucket, object, _, err := common.ParseBucketObject(location)
	if err != nil {
		return fmt.Errorf("parsing location from %q failed: %w", location, err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	f, err := os.Create(destFileName)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %w", object, err)
	}
	defer rc.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("f.Close: %w", err)
	}
	log.Printf("Blob %v downloaded to local file %v\n", object, destFileName)

	return nil
}
