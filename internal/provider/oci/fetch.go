package oci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift-pipelines/tekton-caches/internal/tar"

	"github.com/google/go-containerregistry/pkg/crane"
)

func Fetch(hash, target, folder, expectedDigest string, insecure bool) error {
	cacheImageRef := strings.ReplaceAll(target, "{{hash}}", hash)
	fmt.Fprintf(os.Stderr, "Trying to fetch oci image %s in %s\n", cacheImageRef, folder)

	options := []crane.Option{}
	if insecure {
		options = append(options, crane.Insecure)
	}
	// Try to fetch it (if it exists)
	image, err := crane.Pull(cacheImageRef, options...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		return err
	}

	// Validate digest if expectedDigest was provided
	actualDigest, _ := image.Digest()
	if expectedDigest != "" {
		if actualDigest.String() != expectedDigest {
			return fmt.Errorf("OCI integrity validation failed! expected %s, calculated %s", expectedDigest, actualDigest)
		}
		fmt.Fprintf(os.Stderr, "OCI cache digest validated successfully (%s)\n", actualDigest)
	} else {
		fmt.Fprintf(os.Stderr, "No expected digest provided; skipping validation. %s\n", actualDigest)
	}

	f, err := os.Create(filepath.Join(folder, CacheFile))
	if err != nil {
		return err
	}
	// If it exists, fetch and extract it
	if err := crane.Export(image, f); err != nil {
		return err
	}
	f.Close()

	f, err = os.Open(filepath.Join(folder, CacheFile))
	if err != nil {
		return err
	}
	defer f.Close()
	err = tar.ExtractTar(f, folder)
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	return nil
}
