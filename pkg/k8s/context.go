package k8s

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func SetNamespace(namespace string) error {
	cmd := exec.Command("kubectl", "config", "set-context", "--current", "--namespace", namespace)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetCurrentNamespace() (string, error) {
	cmd := exec.Command("kubectl", "config", "view", "--minify", "-o", "jsonpath='{.contexts[0].context.namespace}'")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Remove surrounding quotes and trim whitespace
	result := strings.Trim(strings.TrimSpace(string(output)), "'\"")
	return result, nil
}

// FindContextForCluster checks if a kubeconfig context already exists for the
// given cluster ARN and that its credentials are still valid.
// Returns the context name and true only when the context exists and a
// lightweight API call (ServerVersion) succeeds.
func FindContextForCluster(clusterARN string) (string, bool) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := loadingRules.Load()
	if err != nil {
		return "", false
	}

	contextName := ""
	for name, ctx := range config.Contexts {
		if ctx.Cluster == clusterARN {
			contextName = name
			break
		}
	}
	if contextName == "" {
		return "", false
	}

	// Build a client targeting the found context and verify the credentials
	// are still valid with a cheap ServerVersion call (no RBAC required).
	overrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	clientConfig := clientcmd.NewDefaultClientConfig(*config, overrides)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return contextName, false
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return contextName, false
	}

	if _, err := clientset.Discovery().ServerVersion(); err != nil {
		return contextName, false
	}

	return contextName, true
}

// UseContext switches the current kubeconfig context.
func UseContext(contextName string) error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := loadingRules.Load()
	if err != nil {
		return err
	}

	if _, exists := config.Contexts[contextName]; !exists {
		return fmt.Errorf("context %q not found", contextName)
	}

	config.CurrentContext = contextName
	return clientcmd.WriteToFile(*config, loadingRules.GetDefaultFilename())
}
