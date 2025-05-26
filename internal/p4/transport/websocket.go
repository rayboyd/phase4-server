// SPDX-License-Identifier: Apache-2.0
package transport

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func NewWebSocketTransport(addr, path string) (*WebSocketTransport, error) {
	wst := &WebSocketTransport{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow all origins for simplicity, adjust for internet facing services.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:     make(map[*websocket.Conn]bool),
		serverAddr:  addr,
		serverPath:  path,
		shutdownSig: make(chan struct{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(path, wst.handleWebSocket)
	wst.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	go func() {
		log.Printf("WebSocketTransport: Starting server on %s%s", addr, path)
		if err := wst.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("WebSocketTransport: HTTP server ListenAndServe error: %v", err)
		}
		log.Printf("WebSocketTransport: Server shut down.")
	}()

	return wst, nil
}

func (wst *WebSocketTransport) SendData(jsonData []byte) error {
	wst.clientsMu.RLock()
	clientsSnapshot := make([]*websocket.Conn, 0, len(wst.clients))
	for conn := range wst.clients {
		clientsSnapshot = append(clientsSnapshot, conn)
	}
	wst.clientsMu.RUnlock()

	if len(clientsSnapshot) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	for _, conn := range clientsSnapshot {
		wg.Add(1)
		go func(c *websocket.Conn, dataToSend []byte) {
			defer wg.Done()
			_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			err := c.WriteMessage(websocket.TextMessage, dataToSend)
			_ = c.SetWriteDeadline(time.Time{})

			if err != nil {
				log.Printf("WebSocketTransport: Write error to %s: %v. Removing client.", c.RemoteAddr(), err)
				wst.clientsMu.Lock()
				if _, ok := wst.clients[c]; ok {
					delete(wst.clients, c)
					_ = c.Close()
				}
				wst.clientsMu.Unlock()
			}
		}(conn, jsonData)
	}
	wg.Wait()

	return nil
}

func (wst *WebSocketTransport) Close() error {
	log.Printf("WebSocketTransport: Shutting down...")
	close(wst.shutdownSig) // Signal background tasks if any were using this.

	// Close all client connections.
	wst.clientsMu.Lock()
	for conn := range wst.clients {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutting down"))
		_ = conn.Close()
		delete(wst.clients, conn) // Remove while iterating safely due to lock.
	}
	wst.clientsMu.Unlock()

	// Graceful shutdown of the HTTP server.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := wst.httpServer.Shutdown(ctx); err != nil {
		log.Printf("WebSocketTransport: HTTP server shutdown error: %v", err)
		return err
	}

	log.Printf("WebSocketTransport: Shutdown complete.")
	return nil
}

func (wst *WebSocketTransport) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wst.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocketTransport: Failed to upgrade connection: %v", err)
		return
	}
	log.Printf("WebSocketTransport: Client connected: %s", conn.RemoteAddr())

	wst.clientsMu.Lock()
	wst.clients[conn] = true
	wst.clientsMu.Unlock()

	go func() {
		defer func() {
			wst.clientsMu.Lock()
			delete(wst.clients, conn)
			wst.clientsMu.Unlock()

			_ = conn.Close()
			log.Printf("WebSocketTransport: Client disconnected: %s", conn.RemoteAddr())
		}()
		for {
			// Detect connection closure. Don't process incoming data here.
			if _, _, err := conn.ReadMessage(); err != nil {
				// Check if it's a normal closure or an unexpected error.
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocketTransport: Read error from %s: %v", conn.RemoteAddr(), err)
				}
				break
			}
		}
	}()
}
