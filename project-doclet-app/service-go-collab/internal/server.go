package collab

import (
	"encoding/json"
	"hash/fnv"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	messageUpdate   = "yjs_update"
	messageSnapshot = "yjs_snapshot"
	messagePresence = "presence"
	messageUserName = "user_name"
)

type Server struct {
	hub    *Hub
	broker *NatsBroker
}

func NewServer(hub *Hub, broker *NatsBroker) *Server {
	return &Server{hub: hub, broker: broker}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ws", s.handleWebsocket)
	return logRequests(mux)
}

func (s *Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	documentID := r.URL.Query().Get("document_id")
	clientID := r.URL.Query().Get("client_id")
	if documentID == "" || clientID == "" {
		http.Error(w, "missing document_id or client_id", http.StatusBadRequest)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		conn:       conn,
		send:       make(chan []byte, 256),
		documentID: documentID,
		clientID:   clientID,
	}

	s.hub.Register(client)
	log.Printf("client %s joined %s", clientID, documentID)

	go client.WritePump()
	s.sendUserNameToClient(client, client.clientID)
	for _, otherID := range s.hub.ClientIDs(documentID) {
		if otherID == client.clientID {
			continue
		}
		s.sendUserNameToClient(client, otherID)
	}
	s.broadcastUserName(client)
	client.ReadPump(func(msg Message) {
		s.handleClientMessage(client, msg)
	})

	s.hub.Unregister(client)
	close(client.send)
	log.Printf("client %s left %s", clientID, documentID)
}

func (s *Server) handleClientMessage(client *Client, msg Message) {
	switch msg.Type {
	case messageUpdate:
		if payload := mustMarshal(msg); payload != nil {
			s.hub.Broadcast(msg.DocumentID, payload, msg.ClientID)
		}
		if s.broker != nil {
			s.broker.Publish(SubjectForDocument(msg.DocumentID, "updates"), msg)
		}
	case messagePresence:
		if payload := mustMarshal(msg); payload != nil {
			s.hub.Broadcast(msg.DocumentID, payload, msg.ClientID)
		}
		if s.broker != nil {
			s.broker.Publish(SubjectForDocument(msg.DocumentID, "presence"), msg)
		}
	case messageSnapshot:
		if s.broker != nil {
			s.broker.Publish(SubjectForDocument(msg.DocumentID, "snapshots"), msg)
		}
	default:
		log.Printf("unknown message type: %s", msg.Type)
	}
}

func (s *Server) sendUserName(client *Client) {
	s.sendUserNameToClient(client, client.clientID)
}

func (s *Server) sendUserNameToClient(client *Client, targetID string) {
	payload, err := json.Marshal(userNameMessage(client.documentID, targetID))
	if err != nil {
		log.Printf("user_name marshal error: %v", err)
		return
	}
	select {
	case client.send <- payload:
	default:
		log.Printf("user_name send dropped for %s", client.clientID)
	}
}

func (s *Server) broadcastUserName(client *Client) {
	payload := mustMarshal(userNameMessage(client.documentID, client.clientID))
	if payload == nil {
		return
	}
	s.hub.Broadcast(client.documentID, payload, client.clientID)
}

func (s *Server) SubscribeNATS() error {
	if s.broker == nil {
		return nil
	}

	if _, err := s.broker.Subscribe("doclet.documents.*.updates", func(msg Message) {
		if payload := mustMarshal(msg); payload != nil {
			s.hub.Broadcast(msg.DocumentID, payload, msg.ClientID)
		}
	}); err != nil {
		return err
	}

	if _, err := s.broker.Subscribe("doclet.documents.*.presence", func(msg Message) {
		if payload := mustMarshal(msg); payload != nil {
			s.hub.Broadcast(msg.DocumentID, payload, msg.ClientID)
		}
	}); err != nil {
		return err
	}

	return nil
}

func mustMarshal(msg Message) []byte {
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return nil
	}
	return payload
}

func nameForClient(clientID string) string {
	adjectives := []string{
		"Brisk", "Calm", "Clever", "Golden", "Mellow",
		"Quick", "Quiet", "Sharp", "Sunny", "Witty",
	}
	nouns := []string{
		"Comet", "Falcon", "Harbor", "Lighthouse", "Meadow",
		"Orchard", "River", "Sparrow", "Summit", "Willow",
	}

	hash := fnv.New32a()
	_, _ = hash.Write([]byte(clientID))
	value := hash.Sum32()
	adj := adjectives[int(value)%len(adjectives)]
	noun := nouns[int(value>>8)%len(nouns)]
	return adj + " " + noun
}

func userNameMessage(documentID, clientID string) Message {
	return Message{
		Type:       messageUserName,
		DocumentID: documentID,
		ClientID:   clientID,
		Payload:    nameForClient(clientID),
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("collab %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
