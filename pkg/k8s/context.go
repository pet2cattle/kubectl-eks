package k8s

import (
	"os"
	"os/exec"
	"strings"
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
