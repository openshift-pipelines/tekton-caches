package upload

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestUploadUnknownScheme(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target string
	}{{
		name:   "unknown scheme",
		target: "ftp://example.com/cache/go",
	}, {
		name:   "no scheme",
		target: "example.com/cache/go",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			digest, err := Upload(context.Background(), "ahash", tc.target, t.TempDir(), false)
			assert.ErrorContains(t, err, "unknown schema")
			assert.Equal(t, digest, "")
		})
	}
}

func TestUploadInvalidTarget(t *testing.T) {
	digest, err := Upload(context.Background(), "ahash", "oci://%zz", t.TempDir(), false)
	assert.Assert(t, err != nil)
	assert.Equal(t, digest, "")
}

func TestUploadHashSubstitutedInTarget(t *testing.T) {
	// The {{hash}} placeholder is replaced before the scheme is resolved, so the
	// error message carries the substituted target.
	digest, err := Upload(context.Background(), "ahash", "ftp://example.com/cache/{{hash}}", t.TempDir(), false)
	assert.ErrorContains(t, err, "ftp://example.com/cache/ahash")
	assert.Equal(t, digest, "")
}
