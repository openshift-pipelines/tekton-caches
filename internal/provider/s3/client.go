package s3

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func createClient(endpoint string, useSSL bool) (*minio.Client, error) {
	// FIXME: decide/define how to get thoses
	var accessKeyID, secretAccessKey string
	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return s3Client, nil
}
