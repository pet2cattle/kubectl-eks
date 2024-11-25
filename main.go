package main

import (
	"fmt"
	"os"

	"github.com/pet2cattle/kubectl-eks/cmd"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error retrieving UserHomeDir")
		os.Exit(1)
	}
	cmd.HomeDir = homeDir

	cmd.Execute()
}
