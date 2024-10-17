//go:build e2e
// +build e2e

package tests

import (
	"context"
	"os"
	"testing"

	"github.com/openshift-pipelines/tekton-caches/tests/client"
	"github.com/openshift-pipelines/tekton-caches/tests/resources"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/yaml"
)

func TestCacheGCS(t *testing.T) {
	ctx := context.Background()
	pr := new(tektonv1.PipelineRun)
	b, err := os.ReadFile("test-pipelineruns/test-pipelinerun-gcs.yaml")
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	if err := yaml.UnmarshalStrict(b, pr); err != nil {
		t.Fatalf("Error unmarshalling: %v", err)
	}

	tc := client.TektonClient(t)

	_ = tc.PipelineRuns(resources.DefaultNamespace).Delete(ctx, pr.GetName(), metav1.DeleteOptions{
		PropagationPolicy: &resources.DeletePolicy,
	})

	if _, err = tc.PipelineRuns(resources.DefaultNamespace).Create(ctx, pr, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Error creating PipelineRun: %v", err)
	}

	if err := resources.WaitForPipelineRun(ctx, tc, pr, resources.Succeed(pr.Name), resources.WaitInterval); err != nil {
		t.Fatalf("Error waiting for PipelineRun to complete: %v", err)
	}
}
