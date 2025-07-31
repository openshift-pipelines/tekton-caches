package fetch

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/openshift-pipelines/tekton-caches/internal/provider/blob"
	"github.com/openshift-pipelines/tekton-caches/internal/provider/oci"
)

func Fetch(ctx context.Context, hash, target, folder string, insecure bool) error {
	// check that folder exists or automatically create it
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.MkdirAll(folder, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
		if err := os.Chmod(folder, os.ModePerm); err != nil {
			return fmt.Errorf("failed to change permissions of folder: %w", err)
		}
	}
	target = strings.ReplaceAll(target, "{{hash}}", hash)
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	source := strings.TrimPrefix(target, u.Scheme+"://")

	switch u.Scheme {
	case "oci":
		return oci.Fetch(hash, source, folder, insecure)
	case "s3", "gs":
		return blob.Fetch(ctx, *u, folder)
	default:
		return fmt.Errorf("unknown schema: %s", target)
	}
}
