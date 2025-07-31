package oci

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openshift-pipelines/tekton-caches/internal/tar"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

const CacheFile = "cache.tar.gz"

func Upload(_ context.Context, hash, target, folder string, insecure bool) error {
	cacheImageRef := strings.ReplaceAll(target, "{{hash}}", hash)
	fmt.Fprintf(os.Stderr, "Upload %s content to oci image %s\n", folder, cacheImageRef)

	// Try to fetch it (if it exists)
	base := empty.Image
	base = mutate.MediaType(base, types.OCIManifestSchema1)
	base = mutate.ConfigMediaType(base, types.OCIConfigJSON)

	file, err := os.CreateTemp("", CacheFile)
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if err := tar.Compress(folder, file.Name()); err != nil {
		return err
	}

	// FIXME: add context as option
	image, err := crane.Append(base, file.Name())
	if err != nil {
		// If not, warn and do not fail
		fmt.Fprintf(os.Stderr, "Warning: %s", err)
		return nil
	}

	options := []crane.Option{}
	if insecure {
		options = append(options, crane.Insecure)
	}
	if err := crane.Push(image, cacheImageRef, options...); err != nil {
		return err
	}

	return nil
}
