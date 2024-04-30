package fetch

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/openshift-pipelines/tekton-caches/internal/hash"
	"github.com/openshift-pipelines/tekton-caches/internal/upload"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
)

var DEFAULT_REG = "oci://127.0.0.1:5000/cache/go"

func TestFetch(t *testing.T) {
	ctx := context.Background()
	regTarget := DEFAULT_REG
	if os.Getenv("TARGET_REGISTRY") != "" {
		regTarget = os.Getenv("TARGET_REGISTRY")
	}
	tmpdir := fs.NewDir(t, t.Name())
	assert.ErrorContains(t, Fetch(ctx, "ahash", regTarget+"notfound", tmpdir.Path(), false), "MANIFEST_UNKNOWN: manifest unknown")
	defer tmpdir.Remove()
	defer env.ChangeWorkingDir(t, tmpdir.Path())()
	assert.NilError(t, os.WriteFile(filepath.Join(tmpdir.Path(), "go.mod"), []byte("module foo/bar/hello.moto"), 0o644))
	hash, err := hash.Compute([]string{filepath.Join(tmpdir.Path(), "go.mod")})
	assert.NilError(t, err)
	assert.NilError(t, upload.Upload(ctx, hash, regTarget, tmpdir.Path(), true))
	assert.NilError(t, Fetch(ctx, hash, regTarget, tmpdir.Path(), false))
	assert.NilError(t, Fetch(ctx, "unknown", regTarget, tmpdir.Path(), false)) // should not error on unkown hash
}
