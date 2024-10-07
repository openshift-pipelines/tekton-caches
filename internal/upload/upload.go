package upload

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/openshift-pipelines/tekton-caches/internal/tar"

	"github.com/openshift-pipelines/tekton-caches/internal/provider/s3"

	"github.com/openshift-pipelines/tekton-caches/internal/provider/gcs"
	"github.com/openshift-pipelines/tekton-caches/internal/provider/oci"
)

func Upload(ctx context.Context, hash, target, folder string, insecure bool) error {
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	newTarget := strings.TrimPrefix(target, u.Scheme+"://")
	newTarget = strings.ReplaceAll(newTarget, "{{hash}}", hash)
	tarFile, err := os.CreateTemp("", "cache.tar")
	if err != nil {
		log.Fatal(err)
	}
	tarName := tarFile.Name()
	if err := tar.Tarit(folder, tarName); err != nil {
		return err
	}
	defer os.Remove(tarName)
	switch u.Scheme {
	case "oci":
		return oci.Upload(ctx, hash, newTarget, folder, insecure)
	case "s3":
		return s3.Upload(ctx, newTarget, tarName)
	case "gs":
		return gcs.Upload(ctx, newTarget, tarName)
	default:
		return fmt.Errorf("unknown schema: %s", target)
	}
}
