package fetch

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/openshift-pipelines/tekton-caches/internal/provider/oci"
)

func Fetch(ctx context.Context, hash, target, folder string) error {
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	newTarget := strings.TrimPrefix(target, u.Scheme+"://")
	switch u.Scheme {
	case "oci":
		return oci.Fetch(ctx, hash, newTarget, folder)
	case "s3":
		return fmt.Errorf("s3 schema not (yet) supported: %s", target)
	case "gcs":
		return fmt.Errorf("gcs schema not (yet) supported: %s", target)
	default:
		return fmt.Errorf("unknown schema: %s", target)
	}
}
