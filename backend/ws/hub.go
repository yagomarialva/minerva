package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for Minerva frontend
	},
}

// Message represents the payload sent over WebSockets
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	clients   map[*websocket.Conn]bool
	Broadcast chan Message
	mutex     sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		Broadcast: make(chan Message),
	}
}

// Run starts the hub loop to listen for broadcasts
func (h *Hub) Run() {
	for {
		select {
		case message := <-h.Broadcast:
			h.mutex.Lock()
			msgBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling ws message: %v", err)
				h.mutex.Unlock()
				continue
			}

			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, msgBytes)
				if err != nil {
					log.Printf("WebSocket send error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}

// ServeWS handles websocket requests from the peer
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()

	log.Println("Frontend client connected to WebSocket")

	// We must read from the connection to process close messages from the client
	go func() {
		defer func() {
			h.mutex.Lock()
			delete(h.clients, conn)
			h.mutex.Unlock()
			conn.Close()
			log.Println("Frontend client disconnected from WebSocket")
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				break
			}
		}
	}()
}
