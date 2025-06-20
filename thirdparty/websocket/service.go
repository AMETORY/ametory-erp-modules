package websocket

import (
	"net/http"

	"gopkg.in/olahol/melody.v1"
)

type WebsocketService struct {
	Client *melody.Melody
}

func NewWebsocketService() *WebsocketService {
	mel := melody.New()
	mel.Config.MaxMessageSize = 2000

	mel.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	return &WebsocketService{
		Client: mel,
	}
}
