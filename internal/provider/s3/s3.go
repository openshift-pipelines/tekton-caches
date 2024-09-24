package s3

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	AwsEnvName    = "AWS_SHARED_CREDENTIALS_FILE"
	AwsConfigFile = "AWS_CONFIG_FILE"
)

func Upload(ctx context.Context, target, filePath string) error {
	log.Printf("S3: Uploading to  %s", target)
	return upload(ctx, target, filePath)
}

func Fetch(ctx context.Context, source, filePath string) error {
	log.Printf("S3: Downloading %s", source)
	return fetch(ctx, source, filePath)
}

func getS3Client(ctx context.Context) *s3.Client {
	credStore := os.Getenv("CRED_STORE")
	if credStore != "" {
		os.Setenv(AwsEnvName, credStore+"/credentials")
		os.Setenv(AwsConfigFile, credStore+"/config")

		log.Printf("Setting %s to %s", AwsEnvName, os.Getenv(AwsEnvName))
		log.Printf("Setting %s to %s", AwsConfigFile, os.Getenv(AwsConfigFile))
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	s3Client := s3.NewFromConfig(cfg)
	return s3Client
}

func fetch(ctx context.Context, source, filePath string) error {
	s3Client := getS3Client(ctx)
	index := strings.Index(source, "/")
	if index == -1 {
		return fmt.Errorf("invalid S3 URL: %s", source)
	}
	bucket := source[:index]
	key := source[index+1:]
	log.Printf("Downloading from S3. FIle: %s, Bucket: %s Key : %s", filePath, bucket, key)

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		log.Fatal("failed to create folder: " + filePath)
		return err
	}

	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Couldn't get object %v:%v. Error: %v", bucket, key, err)
		return err
	}

	defer result.Body.Close()

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer f.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Error: %v\n", key, err)
	}
	_, err = f.Write(body)
	return err
}

func upload(ctx context.Context, target, filePath string) error {
	// Upload the file
	index := strings.Index(target, "/")
	if index == -1 {
		return fmt.Errorf("invalid S3 URL: %s", target)
	}
	bucket := target[:index]
	key := target[index+1:]
	log.Printf("Uploading to S3. Bucket: %s Key : %s", bucket, key)

	s3Client := getS3Client(ctx)
	file, _ := os.Open(filePath)
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	return err
}
