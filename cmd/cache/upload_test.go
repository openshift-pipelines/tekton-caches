package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openshift-pipelines/tekton-caches/internal/flags"
	"gotest.tools/v3/assert"
)

// TestUploadCmdFlagsRegistered guards against reading a flag in RunE that was
// never registered on the command, which makes every invocation fail with
// "flag accessed but not defined".
func TestUploadCmdFlagsRegistered(t *testing.T) {
	cmd := uploadCmd()
	for _, name := range []string{
		flags.TargetFlag,
		flags.DigestFileFlag,
		flags.FolderFlag,
		flags.WorkingDirFlag,
		flags.InsecureFlag,
		flags.NamespaceFlag,
		flags.PatternsFlag,
	} {
		assert.Assert(t, cmd.Flags().Lookup(name) != nil, "flag %q is read by upload but not registered", name)
	}
}

// TestFetchCmdFlagsRegistered is the same guard for the fetch command, which
// shares the flag names with upload.
func TestFetchCmdFlagsRegistered(t *testing.T) {
	cmd := fetchCmd()
	for _, name := range []string{
		flags.SourceFlag,
		flags.FolderFlag,
		flags.WorkingDirFlag,
		flags.DigestFlag,
		flags.InsecureFlag,
		flags.NamespaceFlag,
		flags.PatternsFlag,
	} {
		assert.Assert(t, cmd.Flags().Lookup(name) != nil, "flag %q is read by fetch but not registered", name)
	}
}

func TestUploadCmdNoMatchingFiles(t *testing.T) {
	cmd := uploadCmd()
	cmd.SetOut(os.Stderr)
	cmd.SetArgs([]string{
		"--" + flags.TargetFlag, "oci://example.com/cache/go",
		"--" + flags.FolderFlag, t.TempDir(),
		"--" + flags.WorkingDirFlag, t.TempDir(),
		"--" + flags.DigestFileFlag, filepath.Join(t.TempDir(), "digest.txt"),
		"--" + flags.PatternsFlag, "**/nothing-matches-this",
	})
	err := cmd.Execute()
	assert.ErrorContains(t, err, "didn't match any files")
}

// TestUploadCmdDigestFileOptional makes sure omitting --digest-file is accepted
// (the command must get past flag parsing before it fails on anything else).
func TestUploadCmdDigestFileOptional(t *testing.T) {
	cmd := uploadCmd()
	cmd.SetOut(os.Stderr)
	cmd.SetArgs([]string{
		"--" + flags.TargetFlag, "oci://example.com/cache/go",
		"--" + flags.FolderFlag, t.TempDir(),
		"--" + flags.WorkingDirFlag, t.TempDir(),
		"--" + flags.PatternsFlag, "**/nothing-matches-this",
	})
	err := cmd.Execute()
	assert.ErrorContains(t, err, "didn't match any files")
}
