package hash

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func Compute(files []string) (string, error) {
	hash := sha256.New()
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			return "", err
		}
		defer file.Close()

		if _, err := io.Copy(hash, file); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
