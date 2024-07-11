package utils

import (
	"context"
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
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
	})

	return req.Stream(ctx)
}
