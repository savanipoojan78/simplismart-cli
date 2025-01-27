package main

import (
    // "context"
    "fmt"
    "os"

    "k8s.io/client-go/tools/clientcmd"
    "github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
    // "k8s.io/client-go/tools/clientcmd/api"
)

var ConnectCmd = &cobra.Command{
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

        // Create a slice to hold context names
        contextNames := make([]string, 0, len(config.Contexts))
        for contextName := range config.Contexts {
            contextNames = append(contextNames, contextName)
        }

        // Prompt for context selection
        prompt := promptui.Select{
            Label: "Select a Kubernetes Context",
            Items: contextNames,
        }

        _, selectedContext, err := prompt.Run()
        if err != nil {
            fmt.Printf("Prompt failed %v\n", err)
            return
        }

        fmt.Printf("You selected: %s\n", selectedContext)
        // Set the current context to the selected one
        config.CurrentContext = selectedContext
        err = clientcmd.WriteToFile(*config, kubeconfigPath)
        if err != nil {
            fmt.Printf("Error saving kubeconfig: %v\n", err)
            return
        }

        fmt.Printf("Current context set to: %s\n", selectedContext)
    },
}