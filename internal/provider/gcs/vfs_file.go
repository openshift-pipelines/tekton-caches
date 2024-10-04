package gcs

import (
	"context"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/backend"
	"github.com/c2fo/vfs/v6/backend/gs"
	"github.com/c2fo/vfs/v6/vfssimple"
)

func File(ctx context.Context, file string) (vfs.File, error) {
	bucketAuth := gs.NewFileSystem()
	bucketAuth = bucketAuth.WithContext(ctx)
	backend.Register("gs://", bucketAuth)

	return vfssimple.NewFile(file)
}
