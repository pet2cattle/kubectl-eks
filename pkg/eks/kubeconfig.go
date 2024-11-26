package eks

import (
	"os"
	"os/exec"
)

func UpdateKubeConfig(profile, region, clusterName string) error {
	// using the shell, run aws eks update-kubeconfig --name <clusterName> --region <region> --profile <profile>
	cmd := exec.Command("aws", "eks", "update-kubeconfig", "--name", clusterName, "--region", region, "--profile", profile)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
