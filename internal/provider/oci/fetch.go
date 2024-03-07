// Package oci implements the oci provider
//
// It handles URI such as: oci://
package oci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/openshift-pipelines/tekton-caches/internal/cache"
)

func Fetch(ctx context.Context, hash, target, folder string) error {
	cacheImageRef := strings.ReplaceAll(target, "{{hash}}", hash)
	fmt.Fprintf(os.Stderr, "Trying to fetch oci image %s in %s\n", cacheImageRef, folder)

	// Try to fetch it (if it exists)
	image, err := crane.Pull(cacheImageRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		return err
	}

	f, err := os.Create(filepath.Join(folder, "cache.tar"))
	if err != nil {
		return err
	}
	// If it exists, fetch and extract it
	if err := crane.Export(image, f); err != nil {
		return err
	}
	f.Close()

	return cache.Extract(ctx, folder, "cache.tar")
}
