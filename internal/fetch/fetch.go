package fetch

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/openshift-pipelines/tekton-caches/internal/provider/oci"
)

func Fetch(ctx context.Context, hash, target, folder string, insecure bool) error {
	// check that folder exists or automatically create it
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.MkdirAll(folder, 0755); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
	}
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	newTarget := strings.TrimPrefix(target, u.Scheme+"://")
	switch u.Scheme {
	case "oci":
		return oci.Fetch(ctx, hash, newTarget, folder, insecure)
	case "s3":
		return fmt.Errorf("s3 schema not (yet) supported: %s", target)
	case "gcs":
		return fmt.Errorf("gcs schema not (yet) supported: %s", target)
	default:
		return fmt.Errorf("unknown schema: %s", target)
	}
}
