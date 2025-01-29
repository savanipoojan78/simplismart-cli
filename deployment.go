package main

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var CreateDeploymentCmd = &cobra.Command{
	Use:   "create-deployment",
	Short: "Create a deployment in the Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
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

		// Create Kubernetes client
		config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		// Create or update the deployment
		_ = createDeployment(name, namespace, image, ports, cpuRequest, cpuLimit, ramRequest, ramLimit, clientset)
		// Create Service
		_ = createService(name, namespace, ports, clientset)

		// Create HPA
		if hpaTarget != "" {
			hpa := createHPA(name, namespace, hpaTarget)
			_, err = clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(context.TODO(), hpa, metav1.CreateOptions{})
			if err != nil {
				panic(fmt.Errorf("failed to create HPA: %v", err))
			}
			fmt.Printf("Created HPA %s\n", hpa.Name)
		}

		// Event source would typically be used with KEDA ScaledObjects
		if eventSource != "" {
			fmt.Printf("Event source '%s' could be used for KEDA scaling configuration\n", eventSource)
		}
	},
}

func createDeployment(name, namespace, image string, ports []string, cpuReq, cpuLimit, ramReq, ramLimit string, clientset *kubernetes.Clientset) *appsv1.Deployment {
	// Check if the deployment already exists
	containerPorts := []corev1.ContainerPort{}
	for _, p := range ports {
		port, _ := strconv.ParseInt(p, 10, 32)
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: int32(port),
		})
	}
	existingDeployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Deployment does not exist, create it
			newDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": name},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": name},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  name,
									Image: image,
									Ports: containerPorts,
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse(cpuReq),
											corev1.ResourceMemory: resource.MustParse(ramReq),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse(cpuLimit),
											corev1.ResourceMemory: resource.MustParse(ramLimit),
										},
									},
								},
							},
						},
					},
				},
			}
			_, err = clientset.AppsV1().Deployments(namespace).Create(context.TODO(), newDeployment, metav1.CreateOptions{})
			if err != nil {
				panic(fmt.Errorf("failed to create deployment: %v", err))
			}
			fmt.Printf("Created deployment %s\n", name)
			return newDeployment
		}
		panic(fmt.Errorf("failed to get deployment: %v", err))
	} else {
		// Deployment exists, update it
		existingDeployment.Spec.Template.Spec.Containers[0].Image = image
		existingDeployment.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] = resource.MustParse(cpuReq)
		existingDeployment.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = resource.MustParse(ramReq)
		existingDeployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] = resource.MustParse(cpuLimit)
		existingDeployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = resource.MustParse(ramLimit)
		_, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), existingDeployment, metav1.UpdateOptions{})
		if err != nil {
			panic(fmt.Errorf("failed to update deployment: %v", err))
		}
		fmt.Printf("Updated deployment %s\n", name)
		return existingDeployment
	}
}

func createService(name, namespace string, ports []string, clientset *kubernetes.Clientset) *corev1.Service {
	servicePorts := make([]corev1.ServicePort, 0, len(ports)) // Preallocate slice
	for i, portStr := range ports {
		port, err := strconv.ParseInt(portStr, 10, 32)
		if err != nil {
			panic(fmt.Errorf("invalid port value '%s': %v", portStr, err)) // Handle error
		}
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name: fmt.Sprintf("port-%d", i),
			Port: int32(port),
		})
	}

	// Check if the service already exists
	existingService, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), fmt.Sprintf("%s-service", name), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Service does not exist, create it
			newService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-service", name),
					Namespace: namespace,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{"app": name},
					Ports:    servicePorts,
					Type:     corev1.ServiceTypeClusterIP,
				},
			}
			createdService, err := clientset.CoreV1().Services(namespace).Create(context.TODO(), newService, metav1.CreateOptions{})
			if err != nil {
				panic(fmt.Errorf("failed to create service: %v", err))
			}
			fmt.Printf("Created service %s\n", createdService.Name)
			return createdService
		}
		panic(fmt.Errorf("failed to get service: %v", err))
	}

	// Service exists, patch it
	existingService.Spec.Ports = servicePorts
	updatedService, err := clientset.CoreV1().Services(namespace).Update(context.TODO(), existingService, metav1.UpdateOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to update service: %v", err))
	}
	fmt.Printf("Updated service %s\n", updatedService.Name)
	return updatedService
}

func createHPA(name, namespace, metric string) *autoscalingv2.HorizontalPodAutoscaler {
	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-hpa", name),
			Namespace: namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       name,
			},
			MinReplicas: int32Ptr(1),
			MaxReplicas: 10,
			Metrics: []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: corev1.ResourceName(metric),
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: int32Ptr(50),
						},
					},
				},
			},
		},
	}
}

func int32Ptr(i int32) *int32 { return &i }
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
	CreateDeploymentCmd.MarkFlagRequired("ports")
}
