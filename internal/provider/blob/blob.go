package blob

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

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

// sanitizeLog strips newline/carriage-return characters from s to prevent log injection.
func sanitizeLog(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	return strings.ReplaceAll(s, "\r", "")
}

func Fetch(ctx context.Context, url url.URL, folder, expectedDigest string) error {
	log.Printf("Downloading cache from %s to %s", sanitizeLog(url.String()), sanitizeLog(folder)) //nolint:gosec
	file, err := os.CreateTemp("", cacheFile)
	if err != nil {
		log.Printf("error creating temp file: %s", err)
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

	// Set up SHA-256 hasher and wrap reader in io.TeeReader
	hasher := sha256.New()
	teeReader := io.TeeReader(rc, hasher)
	// 2. Stream download from bucket into temp file while computing SHA-256 on the fly
	_, err = io.Copy(file, teeReader)
	if err != nil {
		log.Printf("error downloading cache: %s", err)
		return err
	}
	// If expectedDigest is provided then Validate the digest against the fetched archive
	actualDigest := getDigest(hasher)
	if expectedDigest != "" && actualDigest != expectedDigest {
		return fmt.Errorf("cache integrity validation failed: expected %s, got %s", expectedDigest, actualDigest)
	}

	// Reset file cursor to beginning before extraction
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

func Upload(ctx context.Context, url url.URL, folder string) (string, error) {
	log.Printf("Uploading cache to %s from %s", sanitizeLog(url.String()), sanitizeLog(folder)) //nolint:gosec
	file, err := os.CreateTemp("", cacheFile)
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	if err := tar.Compress(folder, file.Name()); err != nil {
		log.Printf("error creating tar file: %s", err)
		return "", err
	}

	bucket, err := openBucket(ctx, url.String())
	if err != nil {
		return "", err
	}
	defer clean(bucket)

	writer, err := bucket.NewWriter(ctx, url.Path[1:], nil)
	if err != nil {
		return "", err
	}

	// Stream upload to cloud and compute local SHA-256 simultaneously
	hasher := sha256.New()
	mw := io.MultiWriter(writer, hasher)
	if _, err := io.Copy(mw, file); err != nil {
		return "", fmt.Errorf("failed to upload blob stream: %w", err)
	}

	return getDigest(hasher), writer.Close()
}

func getDigest(hasher hash.Hash) string {
	return "sha256:" + hex.EncodeToString(hasher.Sum(nil))
}
