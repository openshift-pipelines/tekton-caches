//go:build e2e
// +build e2e

package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/openshift-pipelines/tekton-caches/internal/fetch"
	"github.com/openshift-pipelines/tekton-caches/internal/hash"
	"github.com/openshift-pipelines/tekton-caches/internal/upload"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
)

const (
	defaultReg         = "oci://127.0.0.1:5000/cache/go"
	goModContent       = `module foo/bar/hello.moto`
	hashOfGoModContent = "749da1a3a827cde86850743dd2bbf6b65d13497d4b0ecf88d1df7a77ce687f86"
)

func TestOCIUpload(t *testing.T) {
	ctx := context.Background()
	regTarget := defaultReg
	if os.Getenv("TARGET_REGISTRY") != "" {
		regTarget = os.Getenv("TARGET_REGISTRY")
	}
	tmpdir := fs.NewDir(t, t.Name())
	assert.ErrorContains(t, fetch.Fetch(ctx, "ahash", regTarget+"notfound", tmpdir.Path(), false), "MANIFEST_UNKNOWN: manifest unknown")
	defer tmpdir.Remove()
	defer env.ChangeWorkingDir(t, tmpdir.Path())()
	assert.NilError(t, os.WriteFile(filepath.Join(tmpdir.Path(), "go.mod"), []byte(goModContent), 0o644))
	hash, err := hash.Compute([]string{filepath.Join(tmpdir.Path(), "go.mod")})
	assert.Equal(t, hash, hashOfGoModContent)
	assert.NilError(t, err)
	assert.NilError(t, upload.Upload(ctx, hash, regTarget, tmpdir.Path(), true))
	assert.NilError(t, fetch.Fetch(ctx, hash, regTarget, tmpdir.Path(), false))
	assert.NilError(t, fetch.Fetch(ctx, "unknown", regTarget, tmpdir.Path(), false)) // should not error on unknown hash
}
