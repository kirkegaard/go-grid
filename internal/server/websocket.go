package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const THROTTLE = 250

// Upgrader to handle WebSocket requests
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// Client represents a WebSocket connection
type Client struct {
	Id             string
	Conn           *websocket.Conn
	Send           chan []byte
	LastMessage    time.Time
	LastMessageMux sync.Mutex
	ActionCount    int
}

type clientInfo struct {
	Id          string `json:"id"`
	Ip          string `json:"ip"`
	ActionCount int    `json:"actionCount"`
}

// Hub maintains the set of active clients and broadcasts events
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	BlockList  map[string]time.Time
	mu         sync.Mutex
}

var hub = &Hub{
	Clients:    make(map[*Client]bool),
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	BlockList:  make(map[string]time.Time),
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

	clientId := uuid.New().String()

	client := &Client{
		Id:   clientId,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	hub.Register <- client

	// Send the current grid state to the client
	bits, err := getGridState()
	if err != nil {
		log.Printf("Failed to get grid state: %v", err)
		return
	}

	based := base64.StdEncoding.EncodeToString(bits)
	client.Send <- []byte(based)

	go client.readPump()
	go client.writePump()
}

func (h *Hub) GetConnectedClients() []clientInfo {
	h.mu.Lock()
	defer h.mu.Unlock()

	var clientList []clientInfo
	for client := range h.Clients {
		clientList = append(clientList, clientInfo{
			Id:          client.Id,
			Ip:          client.Conn.RemoteAddr().String(),
			ActionCount: client.ActionCount,
		})
	}

	// Sort clients by ActionCount in descending order
	sort.Slice(clientList, func(i, j int) bool {
		return clientList[i].ActionCount > clientList[j].ActionCount
	})

	return clientList
}

func (h *Hub) KickClient(clientId string, duration int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.Clients {
		if client.Id == clientId {
			ip, _, _ := net.SplitHostPort(client.Conn.RemoteAddr().String())
			ip = net.ParseIP(ip).String()
			h.BlockList[ip] = time.Now().Add(time.Duration(duration) * time.Second)

			h.Unregister <- client
			client.Conn.Close()
			return true
		}
	}

	return false
}

func (h *Hub) IsBlocked(ip string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	expiry, exist := h.BlockList[ip]

	if !exist {
		return false
	}

	if time.Now().After(expiry) {
		delete(h.BlockList, ip)
		return false
	}

	return true
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
		c.ActionCount++
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
			ip, _, _ := net.SplitHostPort(client.Conn.RemoteAddr().String())

			isBlocked := h.IsBlocked(ip)

			if isBlocked {
				client.Conn.Close()
				continue
			}
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
