package oci

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	// Set up a fake registry.
	s := httptest.NewServer(registry.New())
	defer s.Close()

	u, err := url.Parse(s.URL)
	assert.NoError(t, err, "Failed to parset he url")

	hash := "d98f9152dc810a4d5dcf737c72d6763c56708c6036b70fceb7947a798f628797"
	target := fmt.Sprintf("%s/test/crane:{{hash}}", u.Host)
	folder := t.TempDir()
	insecure := false

	// Make Dir
	err = os.MkdirAll(folder, os.ModePerm)
	assert.NoError(t, err, "Error creating folder for cache")

	// Upload th the folder to target
	err = Upload(context.Background(), hash, target, folder, insecure)
	assert.NoError(t, err, "Failed to push the image")

	// Fetch The cache from source
	source := target
	err = Fetch(hash, source, folder, insecure)
	assert.NoError(t, err, "Fetch should not return any error")

	cacheFilePath := filepath.Join(folder, "cache.tar")
	_, err = os.Stat(cacheFilePath)
	assert.True(t, os.IsNotExist(err), "Cache tar file should be removed after extraction")
}

func TestFetchImageNotFound(t *testing.T) {
	// Set up a fake registry.
	s := httptest.NewServer(registry.New())
	defer s.Close()

	u, err := url.Parse(s.URL)
	assert.NoError(t, err, "Failed to parset he url")

	hash := "nonexistinghash"
	target := fmt.Sprintf("%s/test/crane:{{hash}}", u.Host)
	folder := t.TempDir()
	insecure := false

	err = Fetch(hash, target, folder, insecure)
	assert.Error(t, err, "Fetch should return an error for nonexistent image")
	assert.True(t,
		containsAny(err.Error(), []string{"NAME_UNKNOWN", "MANIFEST_UNKNOWN"}),
		"Error should indicate that the image manifest or name was not found")
}

func TestFetchInvalidFolder(t *testing.T) {
	// Set up a fake registry.
	s := httptest.NewServer(registry.New())
	defer s.Close()

	u, err := url.Parse(s.URL)
	assert.NoError(t, err, "Failed to parset he url")

	hash := "d98f9152dc810a4d5dcf737c72d6763c56708c6036b70fceb7947a798f628797"
	target := fmt.Sprintf("%s/test/crane:{{hash}}", u.Host)
	img, err := random.Image(1024, 5)
	assert.NoError(t, err, "Failed to create random image")
	err = crane.Push(img, fmt.Sprintf("%s/test/crane:%s", u.Host, hash))
	assert.NoError(t, err, "Failed to push image to registry")

	folder := "/tmp/readonly-dir-for-unit-testing"
	_ = os.MkdirAll(folder, 0o555)
	defer os.RemoveAll(folder)
	insecure := false

	err = Fetch(hash, target, folder, insecure)
	assert.Error(t, err, "Fetch should return an error when folder is not writable")
	assert.Contains(t, err.Error(), "permission denied", "Error should indicate permission issues for the folder")
}

func containsAny(errMsg string, substrs []string) bool {
	for _, substr := range substrs {
		if contains := strings.Contains(errMsg, substr); contains {
			return true
		}
	}
	return false
}
