package cache

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Archive(ctx context.Context, folder, tarball string) (string, error) {
	file, err := os.CreateTemp("", tarball)
	if err != nil {
		return "", err
	}
	tarballPath := file.Name()
	defer os.Remove(tarballPath)
	err = tarit(folder, tarballPath)
	return tarballPath, err
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
