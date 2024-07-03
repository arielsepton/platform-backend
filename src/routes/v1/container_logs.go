package v1

import (
	"github.com/dana-team/platform-backend/src/controllers"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

// 1. receive capp name and namespace (assuming container name is the capp name)
// 2. get pods by label -> rcs.dana.io/parent-capp=capp-sample
// 3. get logs by container name.

func GetPodLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		client, exists := c.Get("kubeClient")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Kubernetes client not found"})
			return
		}

		kubeClient := client.(kubernetes.Interface)
		namespace := c.Param("namespace")
		podName := c.Param("name")
		containerName := c.DefaultQuery("container", podName) // TODO: should default container name be the pod name?

		logs, err := controllers.FetchPodLogs(c.Request.Context(), kubeClient, namespace, podName, containerName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, logs)
	}
}

func GetKnativeServiceLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		client, exists := c.Get("kubeClient")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Kubernetes client not found"})
			return
		}

		kubeClient := client.(kubernetes.Interface)
		ctx := c.Request.Context()

		namespace := c.Param("namespace")
		serviceName := c.Param("name")

		logs, err := controllers.FetchKnativeServiceLogs(ctx, kubeClient, namespace, serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, logs)
	}
}