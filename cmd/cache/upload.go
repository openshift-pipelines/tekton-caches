package main

import (
	"fmt"
	"log"
	"os"

	"github.com/moby/patternmatcher"
	"github.com/openshift-pipelines/tekton-caches/internal/cacheutils"
	"github.com/openshift-pipelines/tekton-caches/internal/flags"
	"github.com/openshift-pipelines/tekton-caches/internal/upload"
	"github.com/spf13/cobra"
)

func uploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "upload",
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := cmd.Flags().GetString(flags.TargetFlag)
			if err != nil {
				return err
			}
			folder, err := cmd.Flags().GetString(flags.FolderFlag)
			if err != nil {
				return err
			}
			workingdir, err := cmd.Flags().GetString(flags.WorkingdirFlag)
			if err != nil {
				return err
			}
			digestFile, err := cmd.Flags().GetString(flags.DigestFileFlag)
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

			insecure, err := cmd.Flags().GetBool(flags.InsecureFlag)
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
			// TODO: use a struct to pas arguments
			digest, err := upload.Upload(cmd.Context(), hashStr, target, folder, insecure)
			if err != nil {
				return err
			}

			// Write the digest of the uploaded cache where the caller asked for it
			if digestFile != "" {
				if err := os.WriteFile(digestFile, []byte(digest), 0o600); err != nil {
					return err
				}
				log.Printf("Successfully wrote digest (%s) to %s\n", digest, digestFile)
			}

			return nil
		},
	}
	cmd.Flags().StringArray(flags.PatternsFlag, []string{}, "Files pattern to compute the hash from")
	cmd.Flags().String(flags.TargetFlag, "", "Cache target reference")
	cmd.Flags().String(flags.FolderFlag, "", "Folder where to extract the content of the cache if it exists")
	cmd.Flags().String(flags.WorkingdirFlag, ".", "Working dir from where the files patterns needs to be taken")
	cmd.Flags().String(flags.NamespaceFlag, "", flags.NamespaceUsage)
	cmd.Flags().String(flags.DigestFileFlag, "", "File to write the digest of the uploaded cache to. If empty, the digest is not written")
	cmd.Flags().Bool(flags.InsecureFlag, false, "Wether to use insecure transport or not to upload to insecure registry")

	return cmd
}
