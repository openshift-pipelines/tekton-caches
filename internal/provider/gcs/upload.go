package gcs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher/pkg/common"
	"github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher/pkg/uploader"
	"google.golang.org/api/option"
)

const (
	userAgent = "tekton-caches"
	// The number of files to upload in parallel.
	workerCount = 200
)

// gcs-uploader -dir examples/ -location gs://tekton-caches-tests/test

func Upload(ctx context.Context, hash, target, folder string) error {
	location := "gs://" + target + hash + ".json"
	client, err := storage.NewClient(ctx, option.WithUserAgent(userAgent))
	if err != nil {
		return fmt.Errorf("failed to create a new gcs client: %w", err)
	}
	bucket, object, generation, err := common.ParseBucketObject(location)
	if err != nil {
		return fmt.Errorf("parsing location from %q failed: %w", location, err)
	}
	if generation != 0 {
		return errors.New("cannot specify manifest file generation")
	}
	_, err = client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil && !errors.Is(err, storage.ErrObjectNotExist) {
		return fmt.Errorf("failed to fetch the object: %w", err)
	}
	// if !errors.Is(err, storage.ErrObjectNotExist) {
	// 	// Delete if the object already exists…
	// 	// It's a workaround to not have the precondition failure…
	// 	if err := client.Bucket(bucket).Object(object).Delete(ctx); err != nil {
	// 		return err
	// 	}
	// }

	u := uploader.New(ctx, realGCS{client, object}, realOS{}, bucket, object, workerCount)

	if err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		return u.Do(ctx, path, info)
	}); err != nil {
		return fmt.Errorf("failed to walk the path: %w", err)
	}

	if err := u.Done(ctx); err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}
	return nil
}
