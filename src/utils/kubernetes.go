package utils

import (
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPodsByLabel receives a kubernetes client, namespace and labelSelector,
// and returns the pods within the namespace with the given label selector
func GetPodsByLabel(ctx context.Context, client kubernetes.Interface, namespace, labelSelector string) (*corev1.PodList, error) {
	return client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
}

// GetPodLogStream receives a kubernetes client, namespace, pod name, and container name,
// and returns the logs of container
func GetPodLogStream(ctx context.Context, client kubernetes.Interface, namespace, podName, containerName string) (io.ReadCloser, error) {
	ok, err := isContainerInPod(ctx, client, namespace, podName, containerName)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("container %v not found in the pod %v", containerName, podName)
	}

	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
		Follow:    true,
		TailLines: &[]int64{int64(10)}[0],
	})

	return req.Stream(ctx)
}

// isContainerInPod checks if a container with the given name exists in the specified pod.
func isContainerInPod(ctx context.Context, client kubernetes.Interface, namespace, podName, containerName string) (bool, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get pod: %v", err)
	}

	if containerName == "" {
		return true, nil
	}

	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return true, nil
		}
	}

	return false, nil
}

// IsPodInPodList checks if a pod with the given name exists in the provided list of pods.
func IsPodInPodList(podName string, pods *corev1.PodList) bool {
	for _, pod := range pods.Items {
		if pod.Name == podName {
			return true // Found the pod by name
		}
	}

	return false
}
