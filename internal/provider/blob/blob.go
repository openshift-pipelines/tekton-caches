package blob

import (
	"context"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/openshift-pipelines/tekton-caches/internal/tar"
	"gocloud.dev/blob"

	// Adding the driver for gcs.
	_ "gocloud.dev/blob/gcsblob"
	// Adding the driver for s3.
	_ "gocloud.dev/blob/s3blob"
	// If we want to add azure blob storage, we can use this import.
	// _ "gocloud.dev/blob/azureblob" .
)

const (
	cacheFile = "cache.tar.gz"
)

var (
	queryParams string
	openBucket  = func(ctx context.Context, urlString string) (*blob.Bucket, error) {
		bucket, err := blob.OpenBucket(ctx, urlString+queryParams)
		return bucket, err
	}
	clean = func(bucket *blob.Bucket) {
		err := bucket.Close()
		if err != nil {
			log.Println("Got error while closing blob")
		}
	}
)

//nolint:gochecknoinits
func init() {
	queryParams = os.Getenv("BLOB_QUERY_PARAMS")
}

func Fetch(ctx context.Context, url url.URL, folder string) error {
	log.Printf("Downloading cache from %s to %s", url.String(), folder)
	file, err := os.CreateTemp("", cacheFile)
	if err != nil {
		log.Printf("error creating tar file: %s", err)
		return err
	}
	defer os.Remove(file.Name())

	bucket, err := openBucket(ctx, url.String())
	if err != nil {
		log.Printf("error opening bucket: %s", err)
		return err
	}
	defer clean(bucket)

	rc, err := bucket.NewReader(ctx, url.Path[1:], nil)
	if err != nil {
		log.Printf("error creating bucket reader: %s", err)
		return err
	}
	defer rc.Close()

	_, err = io.Copy(file, rc)
	if err != nil {
		log.Printf("error downloading cache: %s", err)
		return err
	}
	// Reset cursor to beginning
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	if err := tar.ExtractTarGz(file, folder); err != nil {
		log.Printf("error creating tar file: %s", err)
		return err
	}
	log.Printf("cache untarred %s", folder)
	return nil
}

func Upload(ctx context.Context, url url.URL, folder string) error {
	log.Printf("Uploading cache to %s from %s", url.String(), folder)
	file, err := os.CreateTemp("", cacheFile)
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := tar.Compress(folder, file.Name()); err != nil {
		log.Printf("error creating tar file: %s", err)
		return err
	}

	bucket, err := openBucket(ctx, url.String())
	if err != nil {
		return err
	}
	defer clean(bucket)

	writer, err := bucket.NewWriter(ctx, url.Path[1:], nil)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}

	return writer.Close()
}
