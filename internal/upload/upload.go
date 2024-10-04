package upload

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/openshift-pipelines/tekton-caches/internal/provider/gcs"
	"github.com/openshift-pipelines/tekton-caches/internal/provider/oci"
	"github.com/openshift-pipelines/tekton-caches/internal/provider/s3"
	"github.com/openshift-pipelines/tekton-caches/internal/provider/vfs"
)

func Upload(ctx context.Context, hash, target, folder string, insecure bool) error {
	target = strings.ReplaceAll(target, "{{hash}}", hash)
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	newTarget := strings.TrimPrefix(target, u.Scheme+"://")

	switch u.Scheme {
	case "oci":
		return oci.Upload(ctx, hash, newTarget, folder, insecure)
	case "s3":
		remoteFile, err := s3.File(ctx, target)
		if err != nil {
			return err
		}
		return vfs.Upload(folder, remoteFile)
	case "gs":
		remoteFile, err := gcs.File(ctx, target)
		if err != nil {
			return err
		}
		return vfs.Upload(folder, remoteFile)
	default:
		return fmt.Errorf("unknown schema: %s", target)
	}
}
