package oci

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/stretchr/testify/assert"
)

func TestUpload(t *testing.T) {
	// Step 1: Set up a fake registry
	s := httptest.NewServer(registry.New())
	defer s.Close()

	u, err := url.Parse(s.URL)
	assert.NoError(t, err, "Failed to parse the registry URL")

	hash := "testhash"
	target := fmt.Sprintf("%s/test/crane:{{hash}}", u.Host)
	folder := t.TempDir() // Use a temporary directory as the source folder
	insecure := false

	err = os.WriteFile(fmt.Sprintf("%s/test.txt", folder), []byte("dummy content"), 0o644)
	assert.NoError(t, err, "Failed to create dummy file")

	err = Upload(context.Background(), hash, target, folder, insecure)
	assert.NoError(t, err, "Upload should not return any error")

	pulledImage, err := crane.Pull(fmt.Sprintf("%s/test/crane:testhash", u.Host), crane.Insecure)
	assert.NoError(t, err, "Failed to pull the image back from the registry")

	assert.NotNil(t, pulledImage, "The pulled image should not be nil")

	s.Close()
}
