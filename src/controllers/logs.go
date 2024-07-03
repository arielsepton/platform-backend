package controllers

import (
	"context"
	"fmt"
	"github.com/dana-team/platform-backend/src/utils"
	"io"
	"k8s.io/client-go/kubernetes"
)

const (
	cappLabel = "rcs.dana.io/parent-capp=%s"
)

// FetchPodLogs retrieves the logs of a specific container in a pod.
// It opens a log stream, reads the logs, and returns them as a string.
func FetchPodLogs(ctx context.Context, client kubernetes.Interface, namespace, podName, containerName string) (string, error) {
	logStream, err := utils.GetPodLogStream(ctx, client, namespace, podName, containerName)
	if err != nil {
		return "", fmt.Errorf("error opening log stream: %w", err)
	}
	defer logStream.Close()

	logs, err := io.ReadAll(logStream)
	if err != nil {
		return "", fmt.Errorf("error reading logs: %w", err)
	}

	return string(logs), nil
}

// FetchCappLogs retrieves the logs of a Capp's Knative service.
// It fetches the pods associated with the service, selects the first pod, and retrieves its logs.
func FetchCappLogs(ctx context.Context, client kubernetes.Interface, namespace, cappName, containerName, podName string) (string, error) {
	pods, err := utils.GetPodsByLabel(ctx, client, namespace, fmt.Sprintf(cappLabel, cappName))
	if err != nil {
		return "", fmt.Errorf("error fetching Capp pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for Capp %s in namespace %s", cappName, namespace)
	}

	if podName == "" {
		podName = pods.Items[0].Name
	}

	if containerName == "" {
		containerName = cappName
	}

	return FetchPodLogs(ctx, client, namespace, podName, containerName)
}
