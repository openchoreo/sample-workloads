package collab

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type       string `json:"type"`
	DocumentID string `json:"document_id"`
	ClientID   string `json:"client_id"`
	Payload    string `json:"payload"`
}

type Client struct {
	conn       *websocket.Conn
	send       chan []byte
	documentID string
	clientID   string
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[string]*Client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[string]*Client)}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.documentID] == nil {
		h.clients[client.documentID] = make(map[string]*Client)
	}
	h.clients[client.documentID][client.clientID] = client
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	docClients := h.clients[client.documentID]
	if docClients == nil {
		return
	}
	delete(docClients, client.clientID)
	if len(docClients) == 0 {
		delete(h.clients, client.documentID)
	}
}

func (h *Hub) Broadcast(documentID string, payload []byte, senderID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for clientID, client := range h.clients[documentID] {
		if clientID == senderID {
			continue
		}
		select {
		case client.send <- payload:
		default:
			log.Printf("dropping message for client %s", clientID)
		}
	}
}

func (h *Hub) ClientIDs(documentID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	docClients := h.clients[documentID]
	if docClients == nil {
		return nil
	}
	ids := make([]string, 0, len(docClients))
	for id := range docClients {
		ids = append(ids, id)
	}
	return ids
}

func (c *Client) ReadPump(handle func(Message)) {
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(1024 * 1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("invalid client message: %v", err)
			continue
		}
		msg.DocumentID = c.documentID
		msg.ClientID = c.clientID
		handle(msg)
	}
}

func (c *Client) WritePump() {
	defer c.conn.Close()
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
		}
	}
}
