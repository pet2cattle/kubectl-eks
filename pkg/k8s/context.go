package k8s

import (
	"os"
	"os/exec"
)

func SetNamespace(namespace string) error {

	cmd := exec.Command("kubectl", "config", "set-context", "--current", "--namespace", namespace)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
