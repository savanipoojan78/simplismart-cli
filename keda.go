package main

import (
	// "context"
	"fmt"

	// "k8s.io/client-go/tools/clientcmd"
	"github.com/spf13/cobra"
	// "k8s.io/client-go/tools/clientcmd/api"
)

var installKEDACmd = &cobra.Command{
	Use:   "install-keda",
	Short: "Install KEDA on the Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Command to install KEDA using Helm
		fmt.Println("Installing KEDA using Helm...")
		// Example command: helm install keda kedacore/keda
		// Verify installation and check if KEDA operator is running
	},
}
