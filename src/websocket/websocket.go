package websocket

import (
	"github.com/dana-team/platform-backend/src/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"net/http"
)

type WebSocketHandler interface {
	Register(c *gin.Context) (*websocket.Conn, error)
}

type WebSocket struct {
	upgrader websocket.Upgrader
}

func NewWebSocket(upgrader *websocket.Upgrader) *WebSocket {
	if upgrader == nil {
		upgrader = DefaultUpgrader()
	}

	return &WebSocket{
		upgrader: *upgrader,
	}
}

func (ws *WebSocket) Register(c *gin.Context) (*websocket.Conn, error) {
	ws.upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	token, exists := c.Get("token")
	if !exists {
		return nil, errors.New("Token not found")
	}

	h := http.Header{}
	h.Set(middleware.WebsocketTokenHeader, token.(string))

	conn, err := ws.upgrader.Upgrade(c.Writer, c.Request, h)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
