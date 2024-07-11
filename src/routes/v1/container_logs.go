package v1

import (
	"bufio"
	"fmt"
	"github.com/dana-team/platform-backend/src/controllers"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"time"
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
		// check if empty string
		containerName := c.Param(containerQueryParam)

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func LogsTests() gin.HandlerFunc {
	return func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		creds, exists := c.Get("creds")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Kubernetes client not found"})
			return
		}
		h := http.Header{}
		h.Set("Sec-Websocket-Protocol", creds.(string))

		// client

		kube, exists := c.Get("kubeClient")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Kubernetes client not found"})
			return
		}

		client := kube.(kubernetes.Interface)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, h)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}




		PodLogsConnection := client.CoreV1().Pods("knative-serving-ingress").GetLogs("3scale-kourier-gateway-6b94f8b886-75wdp", &corev1.PodLogOptions{
			Follow:    true,
			TailLines: &[]int64{int64(10)}[0],
		})
		LogStream, _ := PodLogsConnection.Stream(c.Request.Context())
		defer LogStream.Close()
		defer conn.Close()

		reader := bufio.NewScanner(LogStream)
		var line string
		for {
			select {
			case <-c.Done():
				break
			default:
				for reader.Scan() {
					message := fmt.Sprintf("Pod: %v line: %v\n", "3scale-kourier-gateway-6b94f8b886-75wdp", line)
					conn.WriteMessage(websocket.TextMessage, []byte(message))
					time.Sleep(time.Second)
					line = reader.Text()
				}
			}
		}
		// socket
		//conn, err := upgrader.Upgrade(c.Writer, c.Request, h)
		//if err != nil {
		//	c.JSON(500, gin.H{"error": err.Error()})
		//	return
		//}
		//
		//defer conn.Close()
		//for {
		//	conn.WriteMessage(websocket.TextMessage, []byte("Hello Websocket!"))
		//
		//	time.Sleep(time.Second)
		//}
	}
}