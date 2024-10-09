//go:build e2e
// +build e2e

package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektonv1client "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
)

// ConditionAccessorFn is a condition function used polling functions.
type ConditionAccessorFn func(ca apis.ConditionAccessor) (bool, error)

const (
	defaultNamespace = "default"
	waitInterval     = 10 * time.Minute
	tickerDuration   = 10 * time.Second
)

var deletePolicy = metav1.DeletePropagationForeground

func tektonClient(t *testing.T) *tektonv1client.TektonV1Client {
	t.Helper()

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, err := clientConfig.ClientConfig()
	if err != nil {
		t.Fatalf("Error creating client config: %v", err)
	}

	return tektonv1client.NewForConfigOrDie(config)
}

// Succeed provides a poll condition function that checks if the ConditionAccessor
// resource has successfully completed or not.
func Succeed(name string) ConditionAccessorFn {
	return func(ca apis.ConditionAccessor) (bool, error) {
		c := ca.GetCondition(apis.ConditionSucceeded)
		if c != nil {
			if c.Status == corev1.ConditionTrue {
				return true, nil
			} else if c.Status == corev1.ConditionFalse {
				return true, fmt.Errorf("%q failed", name)
			}
		}
		return false, nil
	}
}

func waitForPipelineRun(ctx context.Context, tc *tektonv1client.TektonV1Client, pr *tektonv1.PipelineRun, cond ConditionAccessorFn, duration time.Duration) error {
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()
	stop := make(chan bool)
	done := func() {
		stop <- true
	}
	go func() {
		time.Sleep(duration)
		done()
	}()
	for {
		select {
		case <-stop:
			return fmt.Errorf("timeout waiting for PipelineRun %s to complete", pr.Name)
		case <-ticker.C:
			pr, err := tc.PipelineRuns(defaultNamespace).Get(ctx, pr.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			log.Printf("PipelineRun %s is %v", pr.Name, pr.Status.GetCondition(apis.ConditionSucceeded))
			if val, err := cond(&pr.Status); val && err != nil {
				go done()
				return fmt.Errorf("PipelineRun %s failed: %s", pr.Name, err.Error())
			} else if !val {
				continue
			}
			return nil
		}
	}
}

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

	tc := tektonClient(t)

	_ = tc.PipelineRuns(defaultNamespace).Delete(ctx, pr.GetName(), metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if _, err = tc.PipelineRuns(defaultNamespace).Create(ctx, pr, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Error creating PipelineRun: %v", err)
	}

	if err := waitForPipelineRun(ctx, tc, pr, Succeed(pr.Name), waitInterval); err != nil {
		t.Fatalf("Error waiting for PipelineRun to complete: %v", err)
	}
}
