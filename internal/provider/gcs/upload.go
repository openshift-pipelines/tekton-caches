package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher/pkg/common"
)

// gcs-uploader -dir examples/ -location gs://tekton-caches-tests/test
func Upload(ctx context.Context, target, file string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	// Open local file.
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	location := "gs://" + target
	log.Printf("Uploading %s to %s\n", file, location)
	bucket, object, generation, err := common.ParseBucketObject(location)
	if err != nil {
		return fmt.Errorf("parsing location from %q failed: %w", location, err)
	}
	if generation != 0 {
		return errors.New("cannot specify manifest file generation")
	}
	o := client.Bucket(bucket).Object(object)

	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}
	log.Printf("Blob %v uploaded.\n", object)
	return nil
}
