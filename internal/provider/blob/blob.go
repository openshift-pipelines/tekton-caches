package blob

import (
	"context"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/openshift-pipelines/tekton-caches/internal/tar"
	"gocloud.dev/blob"

	// Adding the driver for gcs
	_ "gocloud.dev/blob/gcsblob"
	// Adding the driver for s3
	_ "gocloud.dev/blob/s3blob"
	// If we want to add azure blob storage, we can use this import
	// _ "gocloud.dev/blob/azureblob"
)

const (
	cacheFile = "cache.tar.gz"
)

func Fetch(ctx context.Context, url url.URL, folder string) error {
	log.Printf("Downloading cache from %s to %s", url.String(), folder)
	file, err := os.CreateTemp("", cacheFile)
	if err != nil {
		log.Printf("error creating tar file: %s", err)
		return err
	}
	defer os.Remove(file.Name())

	bucket, err := blob.OpenBucket(ctx, url.String()+"?usePathStyle=true")
	if err != nil {
		log.Printf("error opening bucket: %s", err)
		return err
	}
	defer bucket.Close()

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

	if err := tar.Untar(ctx, file, folder); err != nil {
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
	if err := tar.Tarit(folder, file.Name()); err != nil {
		log.Printf("error creating tar file: %s", err)
		return err
	}

	bucket, err := blob.OpenBucket(ctx, url.String()+"?usePathStyle=true")
	if err != nil {
		return err
	}
	defer bucket.Close()

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
