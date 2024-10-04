package vfs

import (
	"context"
	"log"
	"os"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/vfssimple"
	"github.com/openshift-pipelines/tekton-caches/internal/tar"
)

func Fetch(ctx context.Context, folder string, remoteFile vfs.File) error {
	file, err := os.CreateTemp("", "cache.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	localFile, err := vfssimple.NewFile("file://" + file.Name())
	if err != nil {
		log.Printf("error creating file: %s", err)
		return err
	}

	err = remoteFile.CopyToFile(localFile)
	if err != nil {
		log.Printf("error copying to location: %s", err)
		return err
	}

	log.Printf("cache downloaded from %s to %s\n", localFile, remoteFile)

	if err := tar.Untar(ctx, file, folder); err != nil {
		log.Printf("error creating tar file: %s", err)
		return err
	}

	log.Printf("cache untarred %s", folder)

	return nil
}
