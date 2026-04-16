package tar //nolint:revive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func Compress(source, target string) error {
	source = filepath.Clean(source)

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

	root, err := os.OpenRoot(source)
	if err != nil {
		return fmt.Errorf("opening source directory: %w", err)
	}
	defer root.Close()

	// Walk using root-scoped paths to avoid symlink TOCTOU (gosec G122).
	err = fs.WalkDir(root.FS(), ".", func(rel string, _ fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walking to %s: %w", rel, walkErr)
		}

		info, err := root.Lstat(rel)
		if err != nil {
			return fmt.Errorf("stat %s: %w", rel, err)
		}

		var linkTarget string
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err = root.Readlink(rel)
			if err != nil {
				return fmt.Errorf("reading symlink: %w", err)
			}
		}

		header, err := tar.FileInfoHeader(info, linkTarget)
		if err != nil {
			return fmt.Errorf("creating tar header: %w", err)
		}

		header.Name = filepath.ToSlash(rel)

		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			header.Uid = int(stat.Uid)
			header.Gid = int(stat.Gid)
		}

		if err := tarball.WriteHeader(header); err != nil {
			return fmt.Errorf("writing header: %w", err)
		}

		if info.Mode().IsRegular() {
			file, err := root.Open(rel)
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			_, copyErr := io.Copy(tarball, file)
			closeErr := file.Close()
			if copyErr != nil {
				return fmt.Errorf("copying file data: %w", copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("closing file: %w", closeErr)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking file tree: %w", err)
	}

	return nil
}

// sanitizeLog strips newline/carriage-return characters from s to prevent log injection.
func sanitizeLog(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	return strings.ReplaceAll(s, "\r", "")
}

func ExtractTarGz(file *os.File, targetDir string) error {
	log.Printf("Extracting tar.gz...%s", sanitizeLog(file.Name())) //nolint:gosec
	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %w", err)
	}

	os.RemoveAll(targetDir)
	defer gz.Close()
	return extract(tar.NewReader(gz), targetDir)
}

func ExtractTar(file *os.File, targetDir string) error {
	log.Printf("Extracting tar...  %s", sanitizeLog(file.Name())) //nolint:gosec
	return extract(tar.NewReader(file), targetDir)
}

func extract(tr *tar.Reader, targetDir string) error {
	baseDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	os.RemoveAll(baseDir)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %w", err)
		}

		// Build the full path
		path, err := safePath(baseDir, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Ensure directory exists; path is validated by safePath above.
			if err := os.MkdirAll(path, os.ModePerm); err != nil { //nolint:gosec
				return err
			}
			// Change the directory permission so that all users can create/update files inside.
			if err := os.Chmod(path, os.ModePerm); path != baseDir && err != nil { //nolint:gosec
				return fmt.Errorf("failed to change permissions of %s: %w", path, err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(path) //nolint:gosec
			if err != nil {
				return fmt.Errorf("error while creating file %s: %w", path, err)
			}

			if err = outFile.Chmod(os.FileMode(0o775)); err != nil {
				return fmt.Errorf("error while updating permissions of the file %s: %w", path, err)
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
			if err = outFile.Close(); err != nil {
				return fmt.Errorf("error closing file %s: %w", path, err)
			}
		case tar.TypeSymlink:
			// Skip Symlinks as they may cause security vulnerability
		default:
			return fmt.Errorf("unsupported file type %v", header.Typeflag)
		}
	}
	return nil
}

func safePath(basePath, targetPath string) (string, error) {
	combinedPath := filepath.Clean(filepath.Join(basePath, targetPath))
	rel, err := filepath.Rel(basePath, combinedPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("target path %q escapes base directory", targetPath)
	}
	return combinedPath, nil
}
