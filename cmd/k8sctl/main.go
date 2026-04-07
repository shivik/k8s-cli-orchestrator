package main

import (
	"fmt"
	"os"

	"k8sctl/pkg/commands"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "k8sctl",
	Short: "A minimal Kubernetes orchestrator CLI",
	Long:  `k8sctl is a lightweight CLI tool for manageing Kubernetes resources with simplified commands.`,
}

func main() {
	rootCmd.AddCommand(commands.DeployCmd())
	rootCmd.AddCommand(commands.StatusCmd())
	rootCmd.AddCommand(commands.LogsCmd())
	rootCmd.AddCommand(commands.DeleteCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
