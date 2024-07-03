package utils

import (
	"context"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPodsByLabel(ctx context.Context, client kubernetes.Interface, namespace, labelSelector string) (*corev1.PodList, error) {
	return client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
}

func GetPodLogStream(ctx context.Context, client kubernetes.Interface, namespace, podName, containerName string) (io.ReadCloser, error) {
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
	})

	// TODO: How to stream logs within HTTP requests
	return req.Stream(ctx)
}
