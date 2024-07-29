package controllers

import (
	"context"
	"fmt"
	"github.com/dana-team/platform-backend/src/types"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ContainerController defines methods to interact with pod containers.
type ContainerController interface {
	GetContainers(namespace, podName string) (types.GetContainersResponse, error)
}

// containerController implements the ContainerController interface.
type containerController struct {
	client kubernetes.Interface
	ctx    context.Context
	logger *zap.Logger
}

// NewContainerController creates a new instance of ContainerController.
func NewContainerController(client kubernetes.Interface, context context.Context, logger *zap.Logger) ContainerController {
	return &containerController{
		client: client,
		ctx:    context,
		logger: logger,
	}
}

// GetContainers returns a list of container names for a given pod in a specific namespace.
func (c *containerController) GetContainers(namespace, podName string) (types.GetContainersResponse, error) {
	c.logger.Debug(fmt.Sprintf("Trying to get all containers in %q namespace", namespace))

	pod, err := c.client.CoreV1().Pods(namespace).Get(c.ctx, podName, metav1.GetOptions{})
	if err != nil {
		c.logger.Error(fmt.Sprintf("failed to get pod %q, in the namespace %q with error: %v", podName, namespace, err.Error()))
		return types.GetContainersResponse{}, &k8serrors.StatusError{ErrStatus: metav1.Status{
			Status:  err.(*k8serrors.StatusError).ErrStatus.Status,
			Code:    err.(*k8serrors.StatusError).ErrStatus.Code,
			Reason:  err.(*k8serrors.StatusError).ErrStatus.Reason,
			Details: err.(*k8serrors.StatusError).ErrStatus.Details,
			Message: fmt.Sprintf("failed to get pod %q, in the namespace %q with error: %s", podName, namespace, err.Error()),
		}}
	}

	response := types.GetContainersResponse{}
	response.Count = len(pod.Spec.Containers)
	for _, container := range pod.Spec.Containers {
		response.Containers = append(
			response.Containers,
			types.Container{
				ContainerName: container.Name,
			})
	}

	c.logger.Debug("Fetched all containers successfully")
	return response, nil
}
