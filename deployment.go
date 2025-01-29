package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
		cpuTarget, _ := cmd.Flags().GetString("cpu-utilization")
		memoryTarget, _ := cmd.Flags().GetString("memory-utilization")

		clientset, err := GetK8sClient()
		if err != nil {
			panic(err.Error())
		}

		// Create or update the deployment
		deployment := createDeployment(name, namespace, image, ports, cpuRequest, cpuLimit, ramRequest, ramLimit, clientset)
		// Create Service
		service := createService(name, namespace, ports, clientset)

		// Create HPA
		err = createScaleObject(name, namespace, cpuTarget, memoryTarget, clientset)
		if err != nil {
			fmt.Printf("Error creating KEDA Scale Object: %v", err)
		}
		// Print deployment and service details
		fmt.Printf("Deployment Name: %s\n", deployment.Name)
		fmt.Printf("Service Name: %s\n", service.Name)
		fmt.Printf("Service IP: %s\n", service.Spec.LoadBalancerIP) // Print service IP
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
		if k8serrors.IsNotFound(err) {
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
		if k8serrors.IsNotFound(err) {
			// Service does not exist, create it
			newService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-service", name),
					Namespace: namespace,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{"app": name},
					Ports:    servicePorts,
					Type:     corev1.ServiceTypeLoadBalancer,
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
func createScaleObject(name, namespace, cpuTarget, memoryTarget string, clientset *kubernetes.Clientset) error {
	scaledObject := map[string]interface{}{
		"apiVersion": "keda.sh/v1alpha1",
		"kind":       "ScaledObject",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"scaleTargetRef": map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"name":       name,
			},
			"pollingInterval": 15,
			"cooldownPeriod":  300,
			"minReplicaCount": 2,
			"maxReplicaCount": 10,
			"triggers": []map[string]interface{}{
				{
					"type": "prometheus",
					"metadata": map[string]interface{}{
						"serverAddress":       "http://prometheus-server.monitoring.svc.cluster.local",
						"query":               fmt.Sprintf(`avg(rate(http_request_duration_seconds_sum{app="%s"}[5m])/rate(http_request_duration_seconds_count{app="%s"}[5m]))`, name, name),
						"threshold":           "0.5",
						"activationThreshold": "0.4",
						"queryValue":          "value",
					},
				},
			},
		},
	}
	if cpuTarget != "" { // Check if cpuTarget is provided
		// Append CPU trigger
		scaledObject["spec"].(map[string]interface{})["triggers"] = append(scaledObject["spec"].(map[string]interface{})["triggers"].([]map[string]interface{}), map[string]interface{}{
			"type":       "cpu",
			"metricType": "Utilization", // Allowed types are 'Utilization' or 'AverageValue'
			"metadata": map[string]interface{}{
				"type":  "Utilization", // Deprecated in favor of trigger.metricType; allowed types are 'Utilization' or 'AverageValue'
				"value": cpuTarget,
			},
		})
	}
	if memoryTarget != "" {
		scaledObject["spec"].(map[string]interface{})["triggers"] = append(scaledObject["spec"].(map[string]interface{})["triggers"].([]map[string]interface{}), map[string]interface{}{
			"type":       "memory",
			"metricType": "Utilization", // Allowed types are 'Utilization' or 'AverageValue'
			"metadata": map[string]interface{}{
				"type":  "Utilization", // Deprecated in favor of trigger.metricType; allowed types are 'Utilization' or 'AverageValue'
				"value": cpuTarget,
			},
		})
	}

	jsonData, err := json.Marshal(scaledObject)
	if err != nil {
		return fmt.Errorf("error marshaling ScaledObject: %v", err)
	}

	// Check if the ScaledObject already exists
	_, err = clientset.RESTClient().
		Get().
		AbsPath("/apis/keda.sh/v1alpha1").
		Namespace(namespace).
		Resource("scaledobjects").
		Name(name).
		DoRaw(context.Background())

	if err != nil {
		if k8serrors.IsNotFound(err) {
			// ScaledObject does not exist, create it
			_, err = clientset.RESTClient().
				Post().
				AbsPath("/apis/keda.sh/v1alpha1").
				Namespace(namespace).
				Resource("scaledobjects").
				Body(jsonData).
				DoRaw(context.Background())

			if err != nil {
				return fmt.Errorf("error creating ScaledObject: %v", err)
			}

			fmt.Printf("Created new ScaledObject: %s", name)
			return nil
		}
		return fmt.Errorf("error checking for existing ScaledObject: %v", err)
	}

	// ScaledObject exists - perform patch
	patchData := map[string]interface{}{
		"spec": scaledObject["spec"],
	}

	jsonPatch, err := json.Marshal(patchData)
	if err != nil {
		return fmt.Errorf("error marshaling patch data: %v", err)
	}

	_, err = clientset.RESTClient().
		Patch(types.MergePatchType).
		AbsPath("/apis/keda.sh/v1alpha1").
		Namespace(namespace).
		Resource("scaledobjects").
		Name(name).
		Body(jsonPatch).
		DoRaw(context.Background())

	if err != nil {
		return fmt.Errorf("error patching ScaledObject: %v", err)
	}
	fmt.Printf("Updated existing ScaledObject: %s", name)
	return nil
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
	CreateDeploymentCmd.Flags().String("cpu-utilization", "", "HPA target metric cpu")
	CreateDeploymentCmd.Flags().String("memory-utilization", "", "HPA target metric memory")
	CreateDeploymentCmd.MarkFlagRequired("image")
	CreateDeploymentCmd.MarkFlagRequired("name")
	CreateDeploymentCmd.MarkFlagRequired("namespace")
	CreateDeploymentCmd.MarkFlagRequired("ports")
}
