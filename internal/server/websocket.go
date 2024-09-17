package server

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const THROTTLE = 500

// Upgrader to handle WebSocket requests
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// Client represents a WebSocket connection
type Client struct {
	Conn           *websocket.Conn
	Send           chan []byte
	LastMessage    time.Time
	LastMessageMux sync.Mutex
}

// Hub maintains the set of active clients and broadcasts events
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

var hub = &Hub{
	Clients:    make(map[*Client]bool),
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}

// GetHub returns the private hub instance
func GetHub() *Hub {
	return hub
}

// webSocketHandler handles WebSocket requests
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open WebSocket connection", http.StatusInternalServerError)
		return
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	hub.Register <- client

	// Send the current grid state to the client
	bits := getGridState()
	compressed := hex.EncodeToString(bits)
	client.Send <- []byte(compressed)

	go client.readPump()
	go client.writePump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	ticker := time.NewTicker(THROTTLE * time.Millisecond)
	defer ticker.Stop()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		c.LastMessageMux.Lock()
		if time.Since(c.LastMessage) < THROTTLE*time.Millisecond {
			c.LastMessageMux.Unlock()
			continue
		}
		c.LastMessage = time.Now()
		c.LastMessageMux.Unlock()

		msg := string(message)
		if strings.HasPrefix(msg, "set:") {
			cellStr := strings.Split(msg, "set:")
			cell, err := strconv.Atoi(cellStr[1])
			if err != nil || cell < 0 || cell >= gridSize {
				log.Printf("Invalid cell: %d", cell)
				continue
			}

			newBit, err := toggleGridCell(cell)
			if err != nil {
				log.Printf("Failed to toggle cell: %d", cell)
				continue
			}

			response := fmt.Sprintf("set:%d:%d", cell, newBit)
			hub.Broadcast <- []byte(response)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer c.Conn.Close()
	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}

	}
}

// RunHub starts the hub to handle broadcasting messages
func (h *Hub) RunHub() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}
