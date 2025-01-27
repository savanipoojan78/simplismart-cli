package main

import (
	// "context"
	"fmt"
	"os"

	// "k8s.io/client-go/tools/clientcmd"
	"github.com/spf13/cobra"
	// "k8s.io/client-go/tools/clientcmd/api"
)

func main() {
	var rootCmd = &cobra.Command{Use: "simplismart-cli"}
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(installKEDACmd)
	rootCmd.AddCommand(createDeploymentCmd)
	rootCmd.AddCommand(healthStatusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
