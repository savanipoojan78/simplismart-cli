package main

import (
    // "context"
    "fmt"
    "os"

    "k8s.io/client-go/tools/clientcmd"
	"github.com/spf13/cobra"
    // "k8s.io/client-go/tools/clientcmd/api"
)

var connectCmd = &cobra.Command{
    Use:   "connect",
    Short: "Connect to the Kubernetes cluster",
    Run: func(cmd *cobra.Command, args []string) {
        kubeconfigPath := os.Getenv("KUBECONFIG")
        if kubeconfigPath == "" {
            kubeconfigPath = clientcmd.RecommendedHomeFile
        }

        config, err := clientcmd.LoadFromFile(kubeconfigPath)
        if err != nil {
            fmt.Printf("Error loading kubeconfig: %v\n", err)
            return
        }

        fmt.Println("Connected to Kubernetes cluster.")
        fmt.Println("Available Kubernetes contexts:")
        for contextName := range config.Contexts {
            fmt.Println(contextName)
        }
    },
}