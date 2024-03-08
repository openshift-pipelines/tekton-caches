package oci

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/openshift-pipelines/tekton-caches/internal/cache"
)

func Upload(ctx context.Context, hash, target, folder string) error {
	cacheImageRef := strings.ReplaceAll(target, "{{hash}}", hash)
	fmt.Fprintf(os.Stderr, "Upload %s content to oci image %s\n", folder, cacheImageRef)

	// Try to fetch it (if it exists)
	base := empty.Image
	base = mutate.MediaType(base, types.OCIManifestSchema1)
	base = mutate.ConfigMediaType(base, types.OCIConfigJSON)

	tarball, err := cache.Archive(ctx, folder, "cache.tar")
	if err != nil {
		return err
	}

	// FIXME: add context as option
	image, err := crane.Append(base, tarball)
	if err != nil {
		// If not, warn and do not fail
		fmt.Fprintf(os.Stderr, "Warning: %s", err)
		return nil
	}

	if err := crane.Push(image, cacheImageRef); err != nil {
		return err
	}

	return nil
}
