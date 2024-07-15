package websocket

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"time"
)

func Stream(c *gin.Context, conn *websocket.Conn, stream io.ReadCloser, formatFunc func(string) string) {
	defer stream.Close()
	defer conn.Close()

	reader := bufio.NewScanner(stream)
	var line string
	for {
		select {
		case <-c.Done():
			return
		default:
			for reader.Scan() {
				line = reader.Text()
				message := formatFunc(line)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					SendErrorMessage(conn, "Error writing message to WebSocket")
					return
				}
				time.Sleep(time.Second)
			}
		}
	}
}

func SendErrorMessage(conn *websocket.Conn, errorMsg string) {
	message := "error: " + errorMsg
	_ = conn.WriteMessage(websocket.TextMessage, []byte(message))
}
