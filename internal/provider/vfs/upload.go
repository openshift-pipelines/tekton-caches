package vfs

import (
	"log"
	"os"

	"github.com/c2fo/vfs/v6"
	"github.com/openshift-pipelines/tekton-caches/internal/tar"

	"github.com/c2fo/vfs/v6/vfssimple"
)

func Upload(folder string, remoteFile vfs.File) error {
	file, err := os.CreateTemp("", "cache.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if err := tar.Tarit(folder, file.Name()); err != nil {
		log.Printf("error creating tar file: %s", err)
		return err
	}

	localFile, err := vfssimple.NewFile("file://" + file.Name())
	if err != nil {
		log.Printf("error creating file: %s", err)
		return err
	}

	err = localFile.CopyToFile(remoteFile)
	if err != nil {
		log.Printf("error copying to location: %s", err)
		return err
	}

	log.Printf("cache uploaded from %s to %s\n", localFile, remoteFile)
	return nil
}
