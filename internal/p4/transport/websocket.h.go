// SPDX-License-Identifier: Apache-2.0
package transport

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketTransport struct {
	clients     map[*websocket.Conn]bool
	httpServer  *http.Server
	shutdownSig chan struct{}
	upgrader    websocket.Upgrader
	serverAddr  string
	serverPath  string
	clientsMu   sync.RWMutex
}
