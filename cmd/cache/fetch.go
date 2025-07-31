package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/moby/patternmatcher"
	"github.com/openshift-pipelines/tekton-caches/internal/fetch"
	"github.com/openshift-pipelines/tekton-caches/internal/flags"
	"github.com/openshift-pipelines/tekton-caches/internal/hash"
	"github.com/spf13/cobra"
)

const (
	workingdirFlag = "workingdir"
	filesFlag      = "hashfiles"
	sourceFlag     = "source"
	folderFlag     = "folder"
	insecureFlag   = "insecure"
)

func fetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "fetch",
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := cmd.Flags().GetString(sourceFlag)
			if err != nil {
				return err
			}
			folder, err := cmd.Flags().GetString(folderFlag)
			if err != nil {
				return err
			}
			workingdir, err := cmd.Flags().GetString(workingdirFlag)
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
					fmt.Fprintf(os.Stderr, "error trying to match files with '%v': %s", patterns, err)
					return false
				}
				return m
			})
			if len(matches) == 0 {
				return fmt.Errorf("didn't match any files with %v", patterns)
			}
			fmt.Fprintf(os.Stderr, "Matched the following files: %v\n", matches)
			// TODO: Hash files based of matches
			hashStr, err := hash.Compute(matches)
			if err != nil {
				return err
			}

			insecure, err := cmd.Flags().GetBool(insecureFlag)
			if err != nil {
				return err
			}
			// FIXME: Wrap the error.
			// If not, warn and do not fail
			// fmt.Fprintf(os.Stderr, "Repository %s doesn't exists or isn't reachable, fetching no cache.\n", cacheImageRef)
			return fetch.Fetch(cmd.Context(), hashStr, target, folder, insecure)
		},
	}

	cmd.Flags().StringArray(flags.PatternsFlag, []string{}, "Files pattern to compute the hash from")
	cmd.Flags().String(sourceFlag, "", "Cache source reference")
	cmd.Flags().String(folderFlag, "", "Folder where to extract the content of the cache if it exists")
	cmd.Flags().String(workingdirFlag, ".", "Working dir from where the files patterns needs to be taken")
	cmd.Flags().Bool(insecureFlag, false, "Wether to use insecure transport or not to upload to insecure registry")

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
		fmt.Fprintf(os.Stderr, "error walking the path %q: %v\n", root, err)
	}
	return files
}
