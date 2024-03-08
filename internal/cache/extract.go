package cache

import (
	"context"
	"os"
	"path/filepath"

	"github.com/codeclysm/extract/v3"
)

func Extract(ctx context.Context, folder, file string) error {
	f, err := os.Open(filepath.Join(folder, "cache.tar"))
	if err != nil {
		return err
	}
	defer f.Close()
	err = extract.Archive(ctx, f, folder, nil)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(folder, "cache.tar")); err != nil {
		return err
	}
	return nil
}
