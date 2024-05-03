package oci

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

func Upload(_ context.Context, hash, target, folder string, insecure bool) error {
	cacheImageRef := strings.ReplaceAll(target, "{{hash}}", hash)
	fmt.Fprintf(os.Stderr, "Upload %s content to oci image %s\n", folder, cacheImageRef)

	// Try to fetch it (if it exists)
	base := empty.Image
	base = mutate.MediaType(base, types.OCIManifestSchema1)
	base = mutate.ConfigMediaType(base, types.OCIConfigJSON)

	file, err := os.CreateTemp("", "cache.tar")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if err := tarit(folder, file.Name()); err != nil {
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

func tarit(source, target string) error {
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			header.Name = strings.TrimPrefix(path, source)

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}
