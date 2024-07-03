package v1

import (
	"github.com/dana-team/platform-backend/src/controllers"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

const (
	namespaceQueryParam = "namespace"
	cappNameQueryParam = "name"
	containerQueryParam = "container"
	podNameQueryParam = "podName"
)

// 1. receive capp name and namespace (assuming container name is the capp name)
// 2. get pods by label -> rcs.dana.io/parent-capp=capp-sample
// 3. get logs by container name.
// 4. tests
// 5. stream
// 6. sockets


// GetPodLogs returns a handler function that fetches logs for a specified pod and container.
// It receives the pod name and namespace from the URL parameters and the container name from the query parameters (defaults to pod name).
func GetPodLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		client, exists := c.Get("kubeClient")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Kubernetes client not found"})
			return
		}

		kubeClient := client.(kubernetes.Interface)
		namespace := c.Param(namespaceQueryParam)
		podName := c.Param(podNameQueryParam)
		containerName := c.DefaultQuery(containerQueryParam, podName)

		logs, err := controllers.FetchPodLogs(c.Request.Context(), kubeClient, namespace, podName, containerName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, logs)
	}
}

// GetCappLogs returns a handler function that fetches logs for a specified Knative service.
// It receives the service name and namespace from the URL parameters and fetches the logs for the first pod associated with the service.
func GetCappLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		client, exists := c.Get("kubeClient")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Kubernetes client not found"})
			return
		}

		kubeClient := client.(kubernetes.Interface)
		ctx := c.Request.Context()

		namespace := c.Param(namespaceQueryParam)
		serviceName := c.Param(cappNameQueryParam)
		containerName := c.DefaultQuery(containerQueryParam, serviceName)
		podName := c.Param(podNameQueryParam)

		logs, err := controllers.FetchCappLogs(ctx, kubeClient, namespace, serviceName, containerName, podName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.String(http.StatusOK, logs)
	}
}