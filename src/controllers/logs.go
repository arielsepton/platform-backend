package controllers

import (
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"

	"github.com/dana-team/platform-backend/src/utils"
	"k8s.io/client-go/kubernetes"
)

const (
	cappLabel = "rcs.dana.io/parent-capp=%s"
)

// FetchPodLogs retrieves the logs of a specific container in a pod.
// It opens a log stream, reads the logs, and returns them as a string.
func FetchPodLogs(ctx context.Context, client kubernetes.Interface, namespace, podName, containerName string) (io.ReadCloser, error) {
	logStream, err := utils.GetPodLogStream(ctx, client, namespace, podName, containerName)
	if err != nil {
		return nil, fmt.Errorf("error opening log stream: %w", err)
	}

	return logStream, nil
}

// FetchCappLogs retrieves the logs of a Capp's Knative service.
// It fetches the pods associated with the service, selects the first pod, and retrieves its logs.
func FetchCappLogs(ctx context.Context, client kubernetes.Interface, namespace, cappName, containerName, podName string) (io.ReadCloser, error) {
	pods, err := utils.GetPodsByLabel(ctx, client, namespace, fmt.Sprintf(cappLabel, cappName))
	if err != nil {
		return nil, fmt.Errorf("error fetching Capp pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for Capp %s in namespace %s", cappName, namespace)
	}

	podName, ok := FetchCappPodName(podName, pods)
	if !ok {
		return nil, fmt.Errorf("pod '%s' not found for Capp %s in namespace %s", podName, cappName, namespace)
	}

	if containerName == "" {
		containerName = cappName
	}

	return FetchPodLogs(ctx, client, namespace, podName, containerName)
}

// FetchCappPodName returns the validated pod name from the provided list of pods.
// If podName is empty, it returns the name of the first pod in the list.
// It also returns a boolean indicating if the pod name was found in the list.
func FetchCappPodName(podName string, pods *corev1.PodList) (string, bool) {
	if podName == "" {
		return pods.Items[0].Name, true
	}

	if utils.IsPodInPodList(podName, pods) {
		return podName, true
	}

	return podName, false
}
