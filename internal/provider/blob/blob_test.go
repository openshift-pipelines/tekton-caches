package blob

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
	digest, err := Upload(ctx, *testURL, tempDir)
	assert.NoError(t, err)
	assert.NotEmpty(t, digest)

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
	if err := Fetch(ctx, *testURL, tempDir, digest); err != nil {
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

func TestFetchAndUpload_InvalidDigest(t *testing.T) {
	expectedDigest := "NonExistentDigest"

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
	digest, err := Upload(ctx, *testURL, tempDir)
	assert.NoError(t, err)
	assert.NotEmpty(t, digest)

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
	err = Fetch(ctx, *testURL, tempDir, expectedDigest)
	assert.EqualError(t, err, fmt.Sprintf("cache integrity validation failed: expected %s, got %s", expectedDigest, digest))
}

// When BLOB_QUERY_PARAMS is not set
// should return empty queryParams.
func TestSanitizeQueryParams_Blank(t *testing.T) {
	sanitizedQueryParams, err := sanitizeQueryParams()
	assert.NoError(t, err)
	assert.Empty(t, sanitizedQueryParams)
}

// When BLOB_QUERY_PARAMS is set to allowed values
// All query params should be retained.
func TestSanitizeQueryParams_Allowed(t *testing.T) {
	t.Setenv("BLOB_QUERY_PARAMS", "fips")
	sanitizedQueryParams, err := sanitizeQueryParams()
	assert.NoError(t, err)
	assert.NotEmpty(t, sanitizedQueryParams)
	assert.Contains(t, sanitizedQueryParams, "fips")
}

// When BLOB_QUERY_PARAMS is set to Disallowed values
// If  forbidden values are set then should return error.
func TestSanitizeQueryParams_Forbidden(t *testing.T) {
	t.Setenv("BLOB_QUERY_PARAMS", "endpoint")
	sanitizedQueryParams, err := sanitizeQueryParams()
	assert.EqualError(t, err, "security policy violation: parameter \"endpoint\" is forbidden")
	assert.Empty(t, sanitizedQueryParams)
}

// When BLOB_QUERY_PARAMS is set to Random values
// When random QueryParams are set then those should be filtered out.
func TestSanitizeQueryParams_Random(t *testing.T) {
	t.Setenv("BLOB_QUERY_PARAMS", "test")
	sanitizedQueryParams, err := sanitizeQueryParams()
	assert.NoError(t, err)
	assert.Empty(t, sanitizedQueryParams)
}
