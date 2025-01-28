package main

import (
	// "context"
	"fmt"

	// "k8s.io/client-go/tools/clientcmd"
	"github.com/spf13/cobra"
	// "k8s.io/client-go/tools/clientcmd/api"
)

var CreateDeploymentCmd = &cobra.Command{
	Use:   "create-deployment",
	Short: "Create a deployment in the Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Use provided details to create a deployment
		fmt.Println("Creating deployment...")
		// Define deployment specifications, service, and HPA
		// Example: kubectl apply -f deployment.yaml
		// Return deployment details
		// Retrieve user inputs
		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")
		namespace, _ := cmd.Flags().GetString("namespace")
		cpuRequest, _ := cmd.Flags().GetString("cpu-request")
		cpuLimit, _ := cmd.Flags().GetString("cpu-limit")
		ramRequest, _ := cmd.Flags().GetString("ram-request")
		ramLimit, _ := cmd.Flags().GetString("ram-limit")
		ports, _ := cmd.Flags().GetStringSlice("ports")
		hpaTarget, _ := cmd.Flags().GetString("hpa-target")
		eventSource, _ := cmd.Flags().GetString("event-source")

		// Use provided details to create a deployment
		fmt.Println("Creating deployment with the following details:")
		fmt.Printf("Name of the Deployment: %s\n", name)
		fmt.Printf("Namespace of the Deployment: %s\n", namespace)
		fmt.Printf("Image: %s\n", image)
		fmt.Printf("CPU Request: %s, CPU Limit: %s\n", cpuRequest, cpuLimit)
		fmt.Printf("RAM Request: %s, RAM Limit: %s\n", ramRequest, ramLimit)
		fmt.Printf("Ports: %v\n", ports)
		fmt.Printf("HPA Target: %s\n", hpaTarget)
		fmt.Printf("Event Source: %s\n", eventSource)
		// Define deployment specifications, service, and HPA
		// Example: kubectl apply -f deployment.yaml
		// Return deployment details
	},
}

func init() {
	CreateDeploymentCmd.Flags().String("name", "", "Name of the deployment")
	CreateDeploymentCmd.Flags().String("image", "", "Docker image and tag (e.g., nginx:latest)")
	CreateDeploymentCmd.Flags().String("namespace", "", "Namespace of the Deployment")
	CreateDeploymentCmd.Flags().String("cpu-request", "100m", "CPU request for the deployment")
	CreateDeploymentCmd.Flags().String("cpu-limit", "500m", "CPU limit for the deployment")
	CreateDeploymentCmd.Flags().String("ram-request", "128Mi", "RAM request for the deployment")
	CreateDeploymentCmd.Flags().String("ram-limit", "512Mi", "RAM limit for the deployment")
	CreateDeploymentCmd.Flags().StringSlice("ports", []string{}, "Ports to expose (e.g., 80,443)")
	CreateDeploymentCmd.Flags().String("hpa-target", "", "HPA target metric (e.g., cpu, memory)")
	CreateDeploymentCmd.Flags().String("event-source", "", "Event source for KEDA metrics (e.g., Kafka, RabbitMQ)")
	CreateDeploymentCmd.MarkFlagRequired("image")
	CreateDeploymentCmd.MarkFlagRequired("name")
	CreateDeploymentCmd.MarkFlagRequired("namespace")
}
