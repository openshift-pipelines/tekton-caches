package cacheutils

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// hashDomain is a domain separator mixed in before anything else. It makes the
// cache keys of this implementation distinct from any other hash computed over
// the same files, and gives us a way to invalidate every key at once should the
// algorithm ever need to change again.
const hashDomain = "tekton-caches/cache-key/v2"

type hashedFile struct {
	// rel is the path of the file relative to the working dir, slash
	// separated. It is what ends up in the hash, so that the key doesn't
	// depend on where the workspace happens to be checked out.
	rel  string
	path string
}

// Compute returns the cache key for files, as a lowercase hex sha256.
//
// Every variable length field (the namespace, the file paths and the file
// contents) is length prefixed before being hashed, so that distinct inputs
// can't be concatenated into the same byte stream: ["AB", "C"] and ["A", "BC"]
// produce different keys.
//
// namespace is an optional tenant salt. When it is empty, pipelines running in
// different namespaces over the same files share a key, which is only safe on a
// cache backend that isn't shared between tenants.
func Compute(namespace, workingdir string, files []string) (string, error) {
	if workingdir == "" {
		workingdir = "."
	}

	hashedFiles := make([]hashedFile, 0, len(files))
	for _, f := range files {
		rel, err := filepath.Rel(workingdir, f)
		if err != nil {
			return "", fmt.Errorf("cannot compute the path of %s relative to %s: %w", f, workingdir, err)
		}
		hashedFiles = append(hashedFiles, hashedFile{rel: filepath.ToSlash(rel), path: f})
	}
	// Don't depend on the order in which the caller matched the files.
	slices.SortFunc(hashedFiles, func(a, b hashedFile) int {
		return strings.Compare(a.rel, b.rel)
	})

	h := sha256.New()
	writeField(h, []byte(hashDomain))
	writeField(h, []byte(namespace))
	writeUint64(h, uint64(len(hashedFiles)))
	for _, f := range hashedFiles {
		writeField(h, []byte(f.rel))
		size, digest, err := hashFile(f.path)
		if err != nil {
			return "", err
		}
		writeUint64(h, uint64(size)) //nolint:gosec // io.Copy never reports a negative count
		// Fixed width, no length prefix needed.
		h.Write(digest)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// hashFile returns the size and the sha256 of the content of path.
func hashFile(path string) (int64, []byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	h := sha256.New()
	size, err := io.Copy(h, file)
	if err != nil {
		return 0, nil, err
	}
	return size, h.Sum(nil), nil
}

// writeField hashes b, prefixed by its length so that the boundary between two
// consecutive fields is unambiguous.
func writeField(h hash.Hash, b []byte) {
	writeUint64(h, uint64(len(b)))
	h.Write(b)
}

func writeUint64(h hash.Hash, n uint64) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], n)
	h.Write(buf[:])
}
