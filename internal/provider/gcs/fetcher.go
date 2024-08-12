package gcs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher/pkg/common"
	"github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher/pkg/fetcher"
	"google.golang.org/api/option"
)

const (
	sourceType    = "Manifest"
	stagingFolder = ".download/"
	backoff       = 100 * time.Millisecond
	retries       = 0
)

func Fetch(ctx context.Context, hash, target, folder string) error {
	location := "gs://" + target + hash + ".json"
	bucket, object, generation, err := common.ParseBucketObject(location)
	if err != nil {
		return fmt.Errorf("parsing location from %q failed: %w", location, err)
	}
	client, err := storage.NewClient(ctx, option.WithUserAgent(userAgent))
	if err != nil {
		return fmt.Errorf("failed to create a new gcs client: %w", err)
	}
	gcs := &fetcher.Fetcher{
		GCS:         realGCS{client, object},
		OS:          realOS{},
		DestDir:     folder,
		StagingDir:  filepath.Join(folder, stagingFolder),
		CreatedDirs: map[string]bool{},
		Bucket:      bucket,
		Object:      object,
		Generation:  generation,
		TimeoutGCS:  true,
		WorkerCount: workerCount,
		Retries:     retries,
		Backoff:     backoff,
		SourceType:  sourceType,
		KeepSource:  false,
		Verbose:     false,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}
	err = gcs.Fetch(ctx)
	if err != nil && !strings.Contains(err.Error(), "storage: object doesn't exist") {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	return nil
}
