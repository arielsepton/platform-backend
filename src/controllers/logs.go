package controllers

import (
	"context"
	"fmt"
	"github.com/dana-team/platform-backend/src/utils"
	"io"
	"k8s.io/client-go/kubernetes"
)

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

func FetchKnativeServiceLogs(ctx context.Context, client kubernetes.Interface, namespace, serviceName string) (string, error) {
	// TODO: should we use this label or the "rcs.dana.io/parent-capp=%s" label?
	pods, err := utils.GetPodsByLabel(ctx, client, namespace, fmt.Sprintf("serving.knative.dev/service=%s", serviceName))
	if err != nil {
		return "", fmt.Errorf("error fetching Knative Service pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for Knative Service %s in namespace %s", serviceName, namespace)
	}

	podName := pods.Items[0].Name
	containerName := serviceName // perhaps it should be configurable
	return FetchPodLogs(ctx, client, namespace, podName, containerName)
}