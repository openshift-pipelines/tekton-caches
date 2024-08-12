package gcs

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
)

// realGCS is a wrapper over the GCS client functions.
type realGCS struct {
	client         *storage.Client
	manifestObject string
}

func (gp realGCS) NewWriter(ctx context.Context, bucket, object string) io.WriteCloser {
	if object != gp.manifestObject {
		object = filepath.Join(".cache", object)
	}
	return gp.client.Bucket(bucket).Object(object).
		If(storage.Conditions{DoesNotExist: true}). // Skip upload if already exists.
		NewWriter(ctx)
}

func (gp realGCS) NewReader(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
	return gp.client.Bucket(bucket).Object(object).NewReader(ctx)
}

// realOS merely wraps the os package implementations.
type realOS struct{}

func (realOS) EvalSymlinks(path string) (string, error)     { return filepath.EvalSymlinks(path) }
func (realOS) Stat(path string) (os.FileInfo, error)        { return os.Stat(path) }
func (realOS) Rename(oldpath, newpath string) error         { return os.Rename(oldpath, newpath) }
func (realOS) Chmod(name string, mode os.FileMode) error    { return os.Chmod(name, mode) }
func (realOS) Create(name string) (*os.File, error)         { return os.Create(name) }
func (realOS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (realOS) Open(name string) (*os.File, error)           { return os.Open(name) }
func (realOS) RemoveAll(path string) error                  { return os.RemoveAll(path) }
