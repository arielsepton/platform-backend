package v1

import (
	"fmt"
	"github.com/dana-team/platform-backend/src/controllers"
	websocketpkg "github.com/dana-team/platform-backend/src/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

const (
	namespaceParam      = "namespace"
	cappNameParam       = "name"
	containerQueryParam = "container"
	podNameQueryParam   = "podName"
)

// GetPodLogs returns a handler function that fetches logs for a specified pod and container.
func GetPodLogs() gin.HandlerFunc {
	return createLogHandler(streamPodLogs, podNameQueryParam, "Pod")
}

// GetCappLogs returns a handler function that fetches logs for a specified Knative service.
func GetCappLogs() gin.HandlerFunc {
	return createLogHandler(streamCappLogs, cappNameParam, "Capp")
}

func createLogHandler(streamFunc func(*gin.Context) (io.ReadCloser, error), paramKey, logPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !websocket.IsWebSocketUpgrade(c.Request) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		websocketClient := websocketpkg.NewWebSocket(nil)
		conn, err := websocketClient.Register(c)
		if err != nil {
			websocketpkg.SendErrorMessage(conn, "Error registering WebSocket")
			return
		}
		defer conn.Close()

		logStream, err := streamFunc(c)
		if err != nil {
			websocketpkg.SendErrorMessage(conn, fmt.Sprintf("Error streaming %v logs: %v", logPrefix, err.Error()))
			return
		}
		defer logStream.Close()

		formatFunc := func(line string) string {
			return fmt.Sprintf("%v: %v line: %v", logPrefix, c.Param(paramKey), line)
		}

		websocketpkg.Stream(c, conn, logStream, formatFunc)
	}
}

func streamPodLogs(c *gin.Context) (io.ReadCloser, error) {
	client, err := getKubeClient(c)
	if err != nil {
		return nil, err
	}

	namespace := c.Param(namespaceParam)
	podName := c.Param(podNameQueryParam)
	containerName := c.Query(containerQueryParam)

	return controllers.FetchPodLogs(c.Request.Context(), client, namespace, podName, containerName)
}

func streamCappLogs(c *gin.Context) (io.ReadCloser, error) {
	client, err := getKubeClient(c)
	if err != nil {
		return nil, err
	}

	namespace := c.Param(namespaceParam)
	cappName := c.Param(cappNameParam)
	containerName := c.DefaultQuery(containerQueryParam, cappName)
	podName := c.Query(podNameQueryParam)

	return controllers.FetchCappLogs(c.Request.Context(), client, namespace, cappName, containerName, podName)
}

func getKubeClient(c *gin.Context) (kubernetes.Interface, error) {
	kube, exists := c.Get("kubeClient")
	if !exists {
		return nil, fmt.Errorf("kube client not found")
	}
	return kube.(kubernetes.Interface), nil
}
