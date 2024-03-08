package s3

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/openshift-pipelines/tekton-caches/internal/cache"
)

func Upload(ctx context.Context, hash, target, folder string) error {
	s3o, err := extractS3Oject(target)
	if err != nil {
		return err
	}
	s3Client, err := createClient(s3o.Endpoint, true)
	if err != nil {
		return err
	}

	tarball, err := cache.Archive(ctx, folder, "cache.tar")
	if err != nil {
		return err
	}

	err = s3Client.MakeBucket(ctx, s3o.Bucket, minio.MakeBucketOptions{Region: s3o.Location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := s3Client.BucketExists(ctx, s3o.Bucket)
		if errBucketExists == nil && exists {
			// log.Printf("We already own %s\n", s3o.Bucket)
		} else {
			return err
		}
	} else {
		// log.Printf("Successfully created %s\n", bucketName)
	}

	contentType := "application/tar"
	// Upload the test file with FPutObject
	_, err = s3Client.FPutObject(ctx, s3o.Bucket, s3o.Object, tarball, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	return nil
}
