package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Hub       *Hub
	Conn      *websocket.Conn
	Send      chan []byte
	closed    bool
	closeMu   sync.Mutex
	closeOnce sync.Once
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
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
				client.closeMu.Lock()
				if !client.closed {
					client.closed = true
					close(client.Send)
				}
				client.closeMu.Unlock()
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

var GlobalHub = NewHub()

func init() {
	go GlobalHub.Run()
}

func BroadcastMessage(typeStr string, payload interface{}) {
	msg := map[string]interface{}{
		"type":    typeStr,
		"payload": payload,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling broadcast message:", err)
		return
	}
	GlobalHub.Broadcast <- jsonMsg
}

func HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to websocket:", err)
		return
	}

	client := &Client{Hub: GlobalHub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	})
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("WebSocket: Write error: %v", err)
				}
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("WebSocket: Ping error: %v", err)
				}
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("WebSocket: Read error: %v", err)
			}
			return
		}

		c.handleMessage(message)
	}
}

func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println("Error unmarshalling message:", err)
		return
	}

	msgType, _ := msg["type"].(string)

	switch msgType {
	case "get_hosts":
		var hosts []models.Host
		if err := db.DB.Preload("Locations").Order("created_at desc").Find(&hosts).Error; err == nil {
			response := map[string]interface{}{
				"type":    "hosts_list",
				"payload": hosts,
			}
			jsonResp, _ := json.Marshal(response)
			select {
			case c.Send <- jsonResp:
			default:
			}
		}
	case "ping":
		response := map[string]interface{}{
			"type":    "pong",
			"payload": nil,
		}
		jsonResp, _ := json.Marshal(response)
		select {
		case c.Send <- jsonResp:
		default:
		}
	}
}

