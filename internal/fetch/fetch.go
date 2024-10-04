package fetch

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/openshift-pipelines/tekton-caches/internal/provider/s3"
	"github.com/openshift-pipelines/tekton-caches/internal/provider/vfs"

	"github.com/openshift-pipelines/tekton-caches/internal/provider/gcs"
	"github.com/openshift-pipelines/tekton-caches/internal/provider/oci"
)

func Fetch(ctx context.Context, hash, target, folder string, insecure bool) error {
	// check that folder exists or automatically create it
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
	}
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	target = strings.ReplaceAll(target, "{{hash}}", hash)
	source := strings.TrimPrefix(target, u.Scheme+"://")

	switch u.Scheme {
	case "oci":
		return oci.Fetch(ctx, hash, source, folder, insecure)
	case "s3":
		remoteFile, err := s3.File(ctx, target)
		if err != nil {
			return err
		}
		return vfs.Fetch(ctx, folder, remoteFile)
	case "gs":
		remoteFile, err := gcs.File(ctx, target)
		if err != nil {
			return err
		}
		return vfs.Fetch(ctx, folder, remoteFile)
	default:
		return fmt.Errorf("unknown schema: %s", target)
	}
}
