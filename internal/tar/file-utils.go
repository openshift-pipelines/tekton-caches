package tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

func Compress(source, target string) error {
	// Create the destination file
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	// Create gzip writer on top of the file
	gz := gzip.NewWriter(tarfile)
	defer gz.Close()

	// Create tar writer on top of the gzip stream
	tarball := tar.NewWriter(gz)
	defer tarball.Close()

	// Walk through the source directory
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking to %s: %w", path, err)
		}

		var linkTarget string
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err = os.Readlink(path)
			if err != nil {
				return fmt.Errorf("reading symlink: %w", err)
			}
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, linkTarget)
		if err != nil {
			return fmt.Errorf("creating tar header: %w", err)
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("getting relative path: %w", err)
		}
		header.Name = relPath

		// Preserve UID/GID
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			header.Uid = int(stat.Uid)
			header.Gid = int(stat.Gid)
		}

		// Write header
		if err := tarball.WriteHeader(header); err != nil {
			return fmt.Errorf("writing header: %w", err)
		}

		// Write file content for regular files only
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer file.Close()

			if _, err := io.Copy(tarball, file); err != nil {
				return fmt.Errorf("copying file data: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking file tree: %w", err)
	}

	return nil
}

func ExtractTarGz(file *os.File, targetDir string) error {
	log.Printf("Extracting tar.gz...%s", file.Name())
	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %w", err)
	}

	os.RemoveAll(targetDir)
	defer gz.Close()
	return extract(tar.NewReader(gz), targetDir)
}

func ExtractTar(file *os.File, targetDir string) error {
	log.Printf("Extracting tar...%s", file.Name())
	return extract(tar.NewReader(file), targetDir)
}

func extract(tr *tar.Reader, targetDir string) error {
	os.RemoveAll(targetDir)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %w", err)
		}

		// Build the full path
		// #nosec
		path := filepath.Join(targetDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			// Ensure directory exists
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
			// Change the directory permission so that all users can create/update files inside.
			if err := os.Chmod(path, os.ModePerm); err != nil {
				return fmt.Errorf("failed to change ownership of %s: %w", path, err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("creating file %s: %w", path, err)
			}
			const maxFileSize = 100 * 1024 * 1024 // 100 MB

			limited := &io.LimitedReader{
				R: tr,
				N: maxFileSize,
			}

			if _, err := io.Copy(outFile, limited); err != nil {
				outFile.Close()
				return fmt.Errorf("error while copying file %s: %w", path, err)
			}
			outFile.Close()
		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, path); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported file type %v", header.Typeflag)
		}
	}
	return nil
}
