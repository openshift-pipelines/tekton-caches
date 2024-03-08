package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/openshift-pipelines/tekton-caches/internal/cache"
)

// accessKeyID := "Q3AM3UQ867SPQQA43P2F"
// secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"

// Initialize minio client object.
func Fetch(ctx context.Context, hash, target, folder string) error {
	s3o, err := extractS3Oject(target)
	if err != nil {
		return err
	}
	s3Client, err := createClient(s3o.Endpoint, true)
	if err != nil {
		return err
	}

	object, err := s3Client.GetObject(context.Background(), s3o.Bucket, s3o.Object, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(folder, "cache.tar"))
	if err != nil {
		return err
	}
	if _, err = io.Copy(f, object); err != nil {
		return err
	}
	f.Close()
	object.Close()

	fmt.Printf("%#v\n", s3Client)
	return cache.Extract(ctx, folder, "cache.tar")
}
