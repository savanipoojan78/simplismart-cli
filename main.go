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
	rootCmd.AddCommand(ConnectCmd)
	rootCmd.AddCommand(InstallKEDACmd)
	rootCmd.AddCommand(CreateDeploymentCmd)
	rootCmd.AddCommand(HealthStatusCmd)
	rootCmd.AddCommand(DoctorCmd) // Added the doctor command
	GenerateDocs(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
