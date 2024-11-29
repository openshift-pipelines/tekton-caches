package tar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeclysm/extract/v4"
)

func Tarit(source, target string) error {
	var buf bytes.Buffer

	// write the .tar.gzip
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	gz := gzip.NewWriter(&buf)
	tarball := tar.NewWriter(gz)

	err = filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if info.IsDir() {
				header.Mode = 0o755 //  create directories with same permissions.
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
	if err != nil {
		return err
	}

	gz.Close()
	tarball.Close()

	if _, err := io.Copy(tarfile, &buf); err != nil {
		log.Println("error copying tarfile", err)
		return err
	}
	return nil
}

func Untar(ctx context.Context, file *os.File, target string) error {
	log.Printf("Untaring file %s to %s", file.Name(), target)
	f, err := os.Open(file.Name())
	if err != nil {
		return err
	}
	defer f.Close()
	return extract.Gz(ctx, f, target, nil)
}
