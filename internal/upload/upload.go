package upload

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
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
	var digest string
	switch u.Scheme {
	case "oci":
		digest, err = oci.Upload(ctx, hash, newTarget, folder, insecure)
	case "s3", "gs":
		digest, err = blob.Upload(ctx, *u, folder)
	default:
		return fmt.Errorf("unknown schema: %s", target)
	}
	if err != nil {
		return err
	}

	// Write the digest to file

	// Write digest to digest.txt with 0644 permissions
	err = os.WriteFile("digest.txt", []byte(digest), fs.ModePerm)
	if err != nil {
		return err
	}
	log.Printf("Successfully wrote digest (%s) to digest.txt\n", digest)
	return nil
}
