package blob

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
)

func TestFetchAndUpload(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "blob_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a mock bucket
	bucket := memblob.OpenBucket(nil)
	defer bucket.Close()

	// Override the openBucket function for testing
	originalOpenBucket := openBucket
	defer func() { openBucket = originalOpenBucket }()
	openBucket = func(_ context.Context, _ string) (*blob.Bucket, error) {
		return bucket, nil
	}
	clean = func(_ *blob.Bucket) {}

	ctx := context.Background()
	testURL, _ := url.Parse("mem://test-bucket/test-object")

	// Test Upload
	if err := Upload(ctx, *testURL, tempDir); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Verify the uploaded content
	data, err := bucket.ReadAll(ctx, "test-object")
	if err != nil {
		t.Fatalf("Failed to read uploaded data: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("Uploaded data is empty")
	}

	// Clean up the temp directory
	os.RemoveAll(tempDir)

	// Create a new temp directory for Fetch
	tempDir, err = os.MkdirTemp("", "blob_test_fetch")
	if err != nil {
		t.Fatalf("Failed to create temp dir for fetch: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test Fetch
	if err := Fetch(ctx, *testURL, tempDir); err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Verify the fetched content
	fetchedFiles, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read fetched directory: %v", err)
	}
	if len(fetchedFiles) == 0 {
		t.Errorf("No files were fetched")
	}
}
