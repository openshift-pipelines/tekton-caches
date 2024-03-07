// Package s3 implements the S3-compatible provider.
//
// It handles URI such as: s3://
// - s3://my-bucket/path/to/something
// - s3://my-bucket/path/to/something?endpoint=s3.amazon.com
// - s3://my-bucket/path/to/something?endpoint=play.minio.io:9000
package s3

import (
	"net/url"
)

type S3Object struct {
	Endpoint string
	Bucket   string
	Object   string
	Location string
}

func extractS3Oject(uri string) (S3Object, error) {
	o := S3Object{}
	url, err := url.Parse(uri)
	if err != nil {
		return o, err
	}
	o.Bucket = url.Host
	o.Object = url.Path

	queries := url.Query()
	if endpoint, ok := queries["endpoint"]; ok {
		o.Endpoint = endpoint[0] // FIXME: handle this
	} else {
		o.Endpoint = "s3.amazon.com"
	}
	if location, ok := queries["location"]; ok {
		o.Location = location[0] // FIXME: handle this
	} else {
		o.Location = "us-east-1"
	}
	return o, nil
}
