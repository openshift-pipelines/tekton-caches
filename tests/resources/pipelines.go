package resources

import (
	"context"
	"fmt"
	"log"
	"time"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektonv1client "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
)

// ConditionAccessorFn is a condition function used polling functions.
type ConditionAccessorFn func(ca apis.ConditionAccessor) (bool, error)

const (
	DefaultNamespace = "default"
	WaitInterval     = 10 * time.Minute
	TickerDuration   = 10 * time.Second
)

var DeletePolicy = metav1.DeletePropagationForeground

// Succeed provides a poll condition function that checks if the ConditionAccessor
// resource has successfully completed or not.
func Succeed(name string) ConditionAccessorFn {
	return func(ca apis.ConditionAccessor) (bool, error) {
		c := ca.GetCondition(apis.ConditionSucceeded)
		if c != nil {
			switch c.Status {
			case corev1.ConditionTrue:
				return true, nil
			case corev1.ConditionFalse:
				return true, fmt.Errorf("%q failed", name)
			case corev1.ConditionUnknown:
			}
		}
		return false, nil
	}
}

func WaitForPipelineRun(ctx context.Context, tc *tektonv1client.TektonV1Client, pr *tektonv1.PipelineRun, cond ConditionAccessorFn, duration time.Duration) error {
	ticker := time.NewTicker(TickerDuration)
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
			pr, err := tc.PipelineRuns(DefaultNamespace).Get(ctx, pr.Name, metav1.GetOptions{})
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
