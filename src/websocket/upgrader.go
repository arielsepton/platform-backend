package websocket

import (
	"github.com/gorilla/websocket"
	"net/http"
)

func DefaultUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			//origin := r.Header.Get("Origin")
			//// Regex, of env var of frontend -> route*
			//return origin == "http://127.0.0.1:8080"
			return true
		},
	}
}
