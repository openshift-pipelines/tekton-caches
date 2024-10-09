package upload

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/openshift-pipelines/tekton-caches/internal/provider/blob"
	"github.com/openshift-pipelines/tekton-caches/internal/provider/oci"
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
	case "s3", "gs":
		return blob.Upload(ctx, *u, folder)
	default:
		return fmt.Errorf("unknown schema: %s", target)
	}
}
