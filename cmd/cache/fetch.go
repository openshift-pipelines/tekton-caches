package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/moby/patternmatcher"
	"github.com/openshift-pipelines/tekton-caches/internal/cacheutils"
	"github.com/openshift-pipelines/tekton-caches/internal/fetch"
	"github.com/openshift-pipelines/tekton-caches/internal/flags"
	"github.com/spf13/cobra"
)

func fetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "fetch",
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := cmd.Flags().GetString(flags.SourceFlag)
			if err != nil {
				return err
			}
			folder, err := cmd.Flags().GetString(flags.FolderFlag)
			if err != nil {
				return err
			}
			workingdir, err := cmd.Flags().GetString(flags.WorkingDirFlag)
			if err != nil {
				return err
			}
			digest, err := cmd.Flags().GetString(flags.DigestFlag)
			if err != nil {
				return err
			}
			namespace, err := cmd.Flags().GetString(flags.NamespaceFlag)
			if err != nil {
				return err
			}
			patterns, err := flags.Patterns(cmd, workingdir)
			if err != nil {
				return err
			}
			matches := glob(workingdir, func(s string) bool {
				m, err := patternmatcher.Matches(s, patterns)
				if err != nil {
					log.Printf("error trying to match files with '%v': %s", patterns, err)
					return false
				}
				return m
			})
			if len(matches) == 0 {
				return fmt.Errorf("didn't match any files with %v", patterns)
			}
			log.Printf("Matched the following files: %v\n", matches)
			hashStr, err := cacheutils.Compute(namespace, workingdir, matches)
			if err != nil {
				return err
			}

			insecure, err := cmd.Flags().GetBool(flags.InsecureFlag)
			if err != nil {
				return err
			}
			// FIXME: Wrap the error.
			// If not, warn and do not fail
			// log.Printf("Repository %s doesn't exists or isn't reachable, fetching no cache.\n", cacheImageRef)
			return fetch.Fetch(cmd.Context(), hashStr, target, folder, digest, insecure)
		},
	}

	cmd.Flags().StringArray(flags.PatternsFlag, []string{}, "Files pattern to compute the hash from")
	cmd.Flags().String(flags.SourceFlag, "", "Cache source reference")
	cmd.Flags().String(flags.FolderFlag, "", "Folder where to extract the content of the cache if it exists")
	cmd.Flags().String(flags.WorkingDirFlag, ".", "Working dir from where the files patterns needs to be taken")
	cmd.Flags().String(flags.DigestFlag, "", "expectedDigest to validate downloaded cache archive")
	cmd.Flags().String(flags.NamespaceFlag, "", flags.NamespaceUsage)
	cmd.Flags().Bool(flags.InsecureFlag, false, "Whether to use insecure transport or not to upload to insecure registry")
	return cmd
}

func glob(root string, fn func(string) bool) []string {
	var files []string
	if err := filepath.WalkDir(root, func(s string, _ fs.DirEntry, _ error) error {
		if fn(s) {
			files = append(files, s)
		}
		return nil
	}); err != nil {
		log.Printf("error walking the path %q: %v\n", root, err)
	}
	return files
}
