package main

import (
	"path/filepath"

	"github.com/openshift-pipelines/tekton-caches/internal/hash"
	"github.com/openshift-pipelines/tekton-caches/internal/upload"
	"github.com/spf13/cobra"
)

func uploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "upload",
		RunE: func(cmd *cobra.Command, args []string) error {
			files, err := cmd.Flags().GetString(filesFlag)
			if err != nil {
				return err
			}
			target, err := cmd.Flags().GetString(targetFlag)
			if err != nil {
				return err
			}
			folder, err := cmd.Flags().GetString(folderFlag)
			if err != nil {
				return err
			}
			// FIXME error out if empty

			matches, err := filepath.Glob(files)
			if err != nil {
				return err
			}
			// TODO: Hash files based of matches
			hashStr, err := hash.Compute(matches)
			if err != nil {
				return err
			}
			return upload.Upload(hashStr, target, folder)
		},
	}
	cmd.Flags().String(filesFlag, "", "Files pattern to compute the hash from")
	cmd.Flags().String(targetFlag, "", "Cache oci image target reference")
	cmd.Flags().String(folderFlag, "", "Folder where to extract the content of the cache if it exists")

	return cmd
}
