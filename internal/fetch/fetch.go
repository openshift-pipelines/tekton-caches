package fetch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeclysm/extract/v3"
	"github.com/google/go-containerregistry/pkg/crane"
)

func Try(hash, target, folder string) error {
	cacheImageRef := strings.ReplaceAll(target, "{{hash}}", hash)
	fmt.Fprintf(os.Stderr, "Trying to fetch oci image %s in %s\n", cacheImageRef, folder)

	// Try to fetch it (if it exists)
	image, err := crane.Pull(cacheImageRef)
	if err != nil {
		// If not, warn and do not fail
		fmt.Fprintf(os.Stderr, "Warning: %s", err)
		return nil
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

	f, err = os.Open(filepath.Join(folder, "cache.tar"))
	if err != nil {
		return err
	}
	defer f.Close()
	err = extract.Archive(context.TODO(), f, folder, nil)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(folder, "cache.tar")); err != nil {
		return err
	}
	return nil
}
