package eks

import (
	"os"
	"os/exec"
)

func UpdateKubeConfig(profile, region, clusterName, kubeConfig string) error {
	var cmd *exec.Cmd

	// using the shell, run aws eks update-kubeconfig --name <clusterName> --region <region> --profile <profile>
	if kubeConfig != "" {
		cmd = exec.Command("aws", "eks", "update-kubeconfig", "--name", clusterName, "--region", region, "--profile", profile, "--kubeconfig", kubeConfig)
	} else {
		cmd = exec.Command("aws", "eks", "update-kubeconfig", "--name", clusterName, "--region", region, "--profile", profile)
	}

	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
