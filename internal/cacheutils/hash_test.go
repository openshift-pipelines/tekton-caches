package cacheutils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openshift-pipelines/tekton-caches/internal/cacheutils"
	"gotest.tools/v3/assert"
)

// writeFiles creates the given files in a new temp dir and returns the dir
// along with the absolute path of each file, in the order they were given.
func writeFiles(t *testing.T, files map[string]string, order ...string) (string, []string) {
	t.Helper()

	dir := t.TempDir()
	paths := make([]string, 0, len(order))
	for _, name := range order {
		path := filepath.Join(dir, filepath.FromSlash(name))
		assert.NilError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NilError(t, os.WriteFile(path, []byte(files[name]), 0o644))
		paths = append(paths, path)
	}
	return dir, paths
}

func compute(t *testing.T, namespace string, files map[string]string, order ...string) string {
	t.Helper()

	dir, paths := writeFiles(t, files, order...)
	hash, err := cacheutils.Compute(namespace, dir, paths)
	assert.NilError(t, err)
	return hash
}

// TestComputeFileBoundaries is the SRVKP-13575 repro: hashing the raw
// concatenation of the file contents made ["AB", "C"] and ["A", "BC"] collide.
func TestComputeFileBoundaries(t *testing.T) {
	first := compute(t, "", map[string]string{"a": "AB", "b": "C"}, "a", "b")
	second := compute(t, "", map[string]string{"a": "A", "b": "BC"}, "a", "b")
	assert.Assert(t, first != second, "files split differently must not share a cache key, got %s for both", first)
}

func TestComputeFileNames(t *testing.T) {
	// Same content, different name.
	first := compute(t, "", map[string]string{"go.sum": "content"}, "go.sum")
	second := compute(t, "", map[string]string{"go.mod": "content"}, "go.mod")
	assert.Assert(t, first != second, "renaming a file must change the cache key, got %s for both", first)

	// Same content, same name, different directory.
	nested := compute(t, "", map[string]string{"sub/go.sum": "content"}, "sub/go.sum")
	assert.Assert(t, first != nested, "moving a file must change the cache key, got %s for both", first)
}

func TestComputeNamespace(t *testing.T) {
	files := map[string]string{"go.mod": "module foo/bar"}

	none := compute(t, "", files, "go.mod")
	nsA := compute(t, "ns-a", files, "go.mod")
	nsB := compute(t, "ns-b", files, "go.mod")

	assert.Assert(t, none != nsA, "a namespace must change the cache key")
	assert.Assert(t, nsA != nsB, "two namespaces must not share a cache key")

	// The namespace is length prefixed, so it can't bleed into the next field.
	assert.Assert(t,
		compute(t, "ab", map[string]string{"c": ""}, "c") != compute(t, "a", map[string]string{"bc": ""}, "bc"),
		"the namespace must not be concatenated with the file paths")
}

func TestComputeIsOrderIndependent(t *testing.T) {
	files := map[string]string{"go.mod": "module foo/bar", "go.sum": "some sums", "sub/go.mod": "module foo/bar/sub"}

	dir, paths := writeFiles(t, files, "go.mod", "go.sum", "sub/go.mod")
	first, err := cacheutils.Compute("", dir, paths)
	assert.NilError(t, err)

	reversed := []string{paths[2], paths[1], paths[0]}
	second, err := cacheutils.Compute("", dir, reversed)
	assert.NilError(t, err)

	assert.Equal(t, first, second, "the cache key must not depend on the order of the matches")
}

func TestComputeIsLocationIndependent(t *testing.T) {
	files := map[string]string{"go.mod": "module foo/bar", "sub/go.sum": "some sums"}

	first := compute(t, "", files, "go.mod", "sub/go.sum")
	second := compute(t, "", files, "go.mod", "sub/go.sum") // a different temp dir
	assert.Equal(t, first, second, "the cache key must not depend on where the workspace is checked out")
}

// TestComputeIsStable pins the algorithm: changing it is a one-time
// invalidation of every existing cache entry, so it should be deliberate.
func TestComputeIsStable(t *testing.T) {
	hash := compute(t, "my-namespace", map[string]string{
		"go.mod":     "module foo/bar/hello.moto",
		"sub/go.sum": "some sums",
	}, "go.mod", "sub/go.sum")
	assert.Equal(t, hash, "f19ceaca2b09a457acd9bc5652f1b1d414a67073d91a5febcd46206355f9ae7e")
}

func TestComputeMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := cacheutils.Compute("", dir, []string{filepath.Join(dir, "nope")})
	assert.ErrorContains(t, err, "no such file or directory")
}

func TestComputeEmptyWorkingDir(t *testing.T) {
	// An empty working dir means the current dir, matching the CLI default.
	dir, paths := writeFiles(t, map[string]string{"go.mod": "module foo/bar"}, "go.mod")
	expected, err := cacheutils.Compute("", dir, paths)
	assert.NilError(t, err)

	t.Chdir(dir)
	hash, err := cacheutils.Compute("", "", []string{"go.mod"})
	assert.NilError(t, err)
	assert.Equal(t, hash, expected)
}
