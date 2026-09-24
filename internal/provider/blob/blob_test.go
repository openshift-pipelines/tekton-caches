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

// When no query param is given
// should return empty queryParams.
func TestSanitizeQueryParams_Blank(t *testing.T) {
	sanitizedQueryParams, err := sanitizeQueryParams(mustParseQuery(t, ""))
	assert.NoError(t, err)
	assert.Empty(t, sanitizedQueryParams)
}

// When query params are allowed
// All query params should be retained with the key spelling the driver expects.
func TestSanitizeQueryParams_Allowed(t *testing.T) {
	sanitizedQueryParams, err := sanitizeQueryParams(mustParseQuery(t, "fips=true&S3FORCEPATHSTYLE=true&sseType=AES256"))
	assert.NoError(t, err)

	assert.Equal(t, 3, len(sanitizedQueryParams))
	assert.Equal(t, "true", sanitizedQueryParams.Get("fips"))
	assert.Equal(t, "true", sanitizedQueryParams.Get("s3ForcePathStyle"))
	assert.Equal(t, "AES256", sanitizedQueryParams.Get("ssetype"))
}

// When a forbidden query param is given
// should return an error.
func TestSanitizeQueryParams_Forbidden(t *testing.T) {
	sanitizedQueryParams, err := sanitizeQueryParams(mustParseQuery(t, "endpoint"))
	assert.EqualError(t, err, "security policy violation: parameter \"endpoint\" is not from allowed list")
	assert.Empty(t, sanitizedQueryParams)
}

func TestBucketUrl_FinalURL(t *testing.T) {
	t.Setenv(EnvBlobQueryParamsKey, "fips=true&sseType=AES256")
	u, err := bucketURL("s3://test-bucket/test-object?S3FORCEPATHSTYLE=false")

	assert.NoError(t, err)
	assert.Equal(t, "s3://test-bucket/test-object?fips=true&s3ForcePathStyle=false&ssetype=AES256", u.String())
}

// Query params present in the URL itself must be sanitized as well, not just the
// ones coming from BLOB_QUERY_PARAMS.
func TestOpenBucket_SanitizesURLQueryParams(t *testing.T) {
	t.Setenv(EnvBlobQueryParamsKey, "")
	bucket, err := openBucket(context.Background(), "s3://test-bucket/test-object?endpoint=http://attacker.example.com")
	assert.EqualError(t, err, "security policy violation: parameter \"endpoint\" is not from allowed list")
	assert.Nil(t, bucket)
}

func TestOpenBucket_InvalidUrl(t *testing.T) {
	t.Setenv(EnvBlobQueryParamsKey, "")
	_, err := openBucket(context.Background(), "s3:/test-bucket/test-object")
	assert.Error(t, err)
}

func TestOpenBucket_InvalidQueryParams(t *testing.T) {
	t.Setenv(EnvBlobQueryParamsKey, "endpoint=http://attacker.example.com;somerandomvalue")
	_, err := openBucket(context.Background(), "s3://test-bucket/test-object")
	assert.Error(t, err)
}

func mustParseQuery(t *testing.T, query string) url.Values {
	t.Helper()
	values, err := url.ParseQuery(query)
	assert.NoError(t, err)
	return values
}
