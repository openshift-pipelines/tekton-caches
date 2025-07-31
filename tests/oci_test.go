//go:build e2e
// +build e2e

package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/openshift-pipelines/tekton-caches/internal/fetch"
	"github.com/openshift-pipelines/tekton-caches/internal/hash"
	"github.com/openshift-pipelines/tekton-caches/internal/upload"
	"github.com/openshift-pipelines/tekton-caches/tests/client"
	"github.com/openshift-pipelines/tekton-caches/tests/resources"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	defaultReg         = "oci://127.0.0.1:5000/cache/go"
	goModContent       = `module foo/bar/hello.moto`
	hashOfGoModContent = "749da1a3a827cde86850743dd2bbf6b65d13497d4b0ecf88d1df7a77ce687f86"
)

func TestOCIUpload(t *testing.T) {
	ctx := context.Background()
	regTarget := defaultReg
	if os.Getenv("TARGET_REGISTRY") != "" {
		regTarget = os.Getenv("TARGET_REGISTRY")
	}
	tmpdir := fs.NewDir(t, t.Name())
	assert.ErrorContains(t, fetch.Fetch(ctx, "ahash", regTarget+"notfound", tmpdir.Path(), false), "MANIFEST_UNKNOWN: manifest unknown")
	defer tmpdir.Remove()
	defer env.ChangeWorkingDir(t, tmpdir.Path())()
	assert.NilError(t, os.WriteFile(filepath.Join(tmpdir.Path(), "go.mod"), []byte(goModContent), 0o644))
	hash, err := hash.Compute([]string{filepath.Join(tmpdir.Path(), "go.mod")})
	assert.Equal(t, hash, hashOfGoModContent)
	assert.NilError(t, err)
	assert.NilError(t, upload.Upload(ctx, hash, regTarget, tmpdir.Path(), true))
	assert.NilError(t, fetch.Fetch(ctx, hash, regTarget, tmpdir.Path(), false))
	assert.NilError(t, fetch.Fetch(ctx, "unknown", regTarget, tmpdir.Path(), false)) // should not error on unknown hash
}

func TestCacheOCI(t *testing.T) {
	ctx := context.Background()
	p := new(tektonv1.Pipeline)
	pr := new(tektonv1.PipelineRun)

	// Get the pipeline yaml
	pipeline, err := os.ReadFile("test-pipelineruns/test-pipeline-oci.yaml")
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	if err := yaml.UnmarshalStrict(pipeline, p); err != nil {
		t.Fatalf("Error unmarshalling: %v", err)
	}

	// Get the pipelineRun yaml
	pipelineRun, err := os.ReadFile("test-pipelineruns/test-pipelinerun-oci.yaml")
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	if err := yaml.UnmarshalStrict(pipelineRun, pr); err != nil {
		t.Fatalf("Error unmarshalling: %v", err)
	}

	tc := client.TektonClient(t)

	// Install the pipeline example
	if p, err = tc.Pipelines(resources.DefaultNamespace).Create(ctx, p, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Error creating Pipeline: %v", err)
	}

	// Install the pipelineRun example
	pr.Spec.PipelineRef = &tektonv1.PipelineRef{Name: p.Name}
	t.Log("Creating a PipelineRun", pr.Spec.PipelineRef.Name)
	if pr, err = tc.PipelineRuns(resources.DefaultNamespace).Create(ctx, pr, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Error creating PipelineRun: %v", err)
	}

	if err := resources.WaitForPipelineRun(ctx, tc, pr, resources.Succeed(pr.Name), resources.WaitInterval); err != nil {
		t.Fatalf("Error waiting for PipelineRun to complete: %v", err)
	}

	// Get the taskrun
	tr, err := tc.TaskRuns(resources.DefaultNamespace).Get(ctx, pr.Name+"-build-task", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error creating PipelineRun: %v", err)
	}

	// Assert the result which was produced by stepAction and stored as result in taskrun
	assert.Equal(t, tr.Status.Results[0].Value.StringVal, "true")

	// Delete the pipelinerun
	err = tc.PipelineRuns(resources.DefaultNamespace).Delete(ctx, pr.GetName(), metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Error deleting pipelinerun")
	}

	// Delete the pipeline
	err = tc.Pipelines(resources.DefaultNamespace).Delete(ctx, p.GetName(), metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Error deleting pipelinerun")
	}
}
