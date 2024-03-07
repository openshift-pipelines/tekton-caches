package s3

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtractS3Object(t *testing.T) {
	testCases := []struct {
		uri    string
		object S3Object
	}{{
		uri: "s3://my-bucket/path/to/something",
		object: S3Object{
			Endpoint: "s3.amazon.com",
			Bucket:   "my-bucket",
			Object:   "/path/to/something",
			Location: "us-east-1",
		},
	}, {
		uri: "s3://my-bucket/path/to/something?endpoint=s3.amazon.com",
		object: S3Object{
			Endpoint: "s3.amazon.com",
			Bucket:   "my-bucket",
			Object:   "/path/to/something",
			Location: "us-east-1",
		},
	}, {
		uri: "s3://my-bucket/path/to/something?location=eu-west-3",
		object: S3Object{
			Endpoint: "s3.amazon.com",
			Bucket:   "my-bucket",
			Object:   "/path/to/something",
			Location: "eu-west-3",
		},
	}, {
		uri: "s3://my-bucket/path/to/something?endpoint=s3.amazon.com&location=eu-west-3",
		object: S3Object{
			Endpoint: "s3.amazon.com",
			Bucket:   "my-bucket",
			Object:   "/path/to/something",
			Location: "eu-west-3",
		},
	}, {
		uri: "s3://my-bucket/path/to/something?endpoint=play.minio.io:9000",
		object: S3Object{
			Endpoint: "play.minio.io:9000",
			Bucket:   "my-bucket",
			Object:   "/path/to/something",
			Location: "us-east-1",
		},
	}}
	for _, tc := range testCases {
		o, err := extractS3Oject(tc.uri)
		if err != nil {
			t.Error(err)
		}
		if d := cmp.Diff(tc.object, o); d != "" {
			t.Errorf("Diff %s", fmt.Sprintf("(-want, +got): %s", d))
		}
	}
}
