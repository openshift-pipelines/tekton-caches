package s3

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/c2fo/vfs/v6"

	"github.com/c2fo/vfs/v6/vfssimple"

	"github.com/c2fo/vfs/v6/backend"
	vfss3 "github.com/c2fo/vfs/v6/backend/s3"
)

func getS3Client(ctx context.Context) *s3.Client {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return s3Client
}

func File(ctx context.Context, file string) (vfs.File, error) {
	bucketAuth := vfss3.NewFileSystem().WithClient(getS3Client(ctx)).WithOptions(vfss3.Options{
		ForcePathStyle: true,
		Endpoint:       os.Getenv("AWS_ENDPOINT_URL"),
	})

	backend.Register("s3://", bucketAuth)

	log.Printf("setting location: %s", file)
	return vfssimple.NewFile(file)
}
