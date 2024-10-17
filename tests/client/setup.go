package client

import (
	"testing"

	tektonv1client "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func TektonClient(t *testing.T) *tektonv1client.TektonV1Client {
	t.Helper()

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, err := clientConfig.ClientConfig()
	if err != nil {
		t.Fatalf("Error creating client config: %v", err)
	}

	return tektonv1client.NewForConfigOrDie(config)
}
