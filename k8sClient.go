package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

func GetK8sClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	}

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func GetMetricsClient() (*metricsv1beta1.Clientset, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	}

	if err != nil {
		return nil, err
	}

	return metricsv1beta1.NewForConfig(config)
}