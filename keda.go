package main

import (
	"context"
	"fmt"

	// "k8s.io/client-go/tools/clientcmd"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// "k8s.io/client-go/tools/clientcmd/api"
)

var InstallKEDACmd = &cobra.Command{
	Use:   "install-keda",
	Short: "Install KEDA on the Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Create a Kubernetes client from the current context
		config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			fmt.Println("Error creating Kubernetes client config:", err)
			return
		}
		// Create a Kubernetes client
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Println("Error creating Kubernetes client:", err)
			return
		}

		// Check if KEDA deployment exists
		deploymentName := "keda-operator"
		namespace := "keda"
		_, err = clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			fmt.Println("KEDA is not running, installing...")
			// Code to install KEDA using Helm goes here
		} else {
			// Check if the KEDA operator pods are running
			pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
				LabelSelector: "app=keda-operator", // Adjust the label selector as needed
			})
			if err != nil {
				fmt.Println("Error retrieving KEDA operator pods:", err)
				return
			}
			if len(pods.Items) == 0 {
				fmt.Println("KEDA operator pods are not running.")
			} else {
				fmt.Println("KEDA operator pods are running.")
			}
		}
	},
}
