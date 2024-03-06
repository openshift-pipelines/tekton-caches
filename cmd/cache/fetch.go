package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/moby/patternmatcher"
	"github.com/openshift-pipelines/tekton-caches/internal/fetch"
	"github.com/openshift-pipelines/tekton-caches/internal/hash"
	"github.com/spf13/cobra"
)

const (
	workingdirFlag = "workingdir"
	filesFlag      = "hashfiles"
	patternsFlag   = "pattern"
	sourceFlag     = "source"
	folderFlag     = "folder"
)

func fetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "fetch",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			patterns, err := cmd.Flags().GetStringArray(patternsFlag)
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
				return fmt.Errorf("Didn't match any files with %v", patterns)
			} else {
				fmt.Fprintf(os.Stderr, "Matched the following files: %v\n", matches)
			}
			// TODO: Hash files based of matches
			hashStr, err := hash.Compute(matches)
			if err != nil {
				return err
			}

			// FIXME: Wrap the error.
			// If not, warn and do not fail
			// fmt.Fprintf(os.Stderr, "Repository %s doesn't exists or isn't reachable, fetching no cache.\n", cacheImageRef)
			return fetch.Fetch(cmd.Context(), hashStr, target, folder)
		},
	}

	cmd.Flags().StringArray(patternsFlag, []string{}, "Files pattern to compute the hash from")
	cmd.Flags().String(sourceFlag, "", "Cache source reference")
	cmd.Flags().String(folderFlag, "", "Folder where to extract the content of the cache if it exists")
	cmd.Flags().String(workingdirFlag, ".", "Working dir from where the files patterns needs to be taken")

	return cmd
}

func glob(root string, fn func(string) bool) []string {
	var files []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if fn(s) {
			files = append(files, s)
		}
		return nil
	})
	return files
}
