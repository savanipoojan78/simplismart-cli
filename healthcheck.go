package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var kubeconfig string

var HealthStatusCmd = &cobra.Command{
	Use:   "health-status",
	Short: "Retrieve health status of a deployment",
	Run: func(cmd *cobra.Command, args []string) {
		deploymentName, _ := cmd.Flags().GetString("name")
		namespace, _ := cmd.Flags().GetString("namespace")
		clientset, err := GetK8sClient()
		if err != nil {
			log.Fatalf("Failed to create Kubernetes client: %v", err)
		}
		metricsClient, err := GetMetricsClient()
		if err != nil {
			log.Fatalf("Failed to create Metrics client: %v", err)
		}

		// Get deployment status
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Failed to get deployment: %v", err)
		}
		fmt.Printf("Deployment: %s, Available Replicas: %d/%d\n", deployment.Name, deployment.Status.AvailableReplicas, *deployment.Spec.Replicas)

		// Get pod status and resource usage
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", deploymentName),
		})
		if err != nil {
			log.Fatalf("Failed to list pods: %v", err)
		}

		for _, pod := range pods.Items {
			fmt.Printf("Pod: %s, Status: %s\n", pod.Name, pod.Status.Phase)
			if pod.Status.Phase != "Running" {
				fmt.Printf("\tWarning: Pod %s is in %s state!\n", pod.Name, pod.Status.Phase)
			}
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					fmt.Printf("\tContainer %s is not ready\n", containerStatus.Name)
				}
			}
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					fmt.Printf("\tContainer %s is not ready\n", containerStatus.Name)
				}
			}

			// Get pod metrics
			podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			if err != nil {
				log.Printf("Failed to get metrics for pod %s: %v", pod.Name, err)
				continue
			}
			for _, container := range podMetrics.Containers {
				fmt.Printf("\tContainer: %s, CPU: %s, Memory: %s\n", container.Name, container.Usage.Cpu().String(), container.Usage.Memory().String())
			}
		}
	},
}

func init() {
	HealthStatusCmd.Flags().String("name", "", "Name of the deployment")
	HealthStatusCmd.Flags().String("namespace", "default", "Namespace of the deployment")
	HealthStatusCmd.MarkFlagRequired("name")
	HealthStatusCmd.MarkFlagRequired("namespace")
}
