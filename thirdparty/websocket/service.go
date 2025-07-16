package websocket

import (
	"net/http"

	"gopkg.in/olahol/melody.v1"
)

type WebsocketService struct {
	Client *melody.Melody
}

// NewWebsocketService creates a new WebsocketService instance with a maximum message size of 2000 bytes.
// It also sets the Upgrader to allow cross-origin requests.
func NewWebsocketService() *WebsocketService {
	mel := melody.New()
	mel.Config.MaxMessageSize = 2000

	mel.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	return &WebsocketService{
		Client: mel,
	}
}
