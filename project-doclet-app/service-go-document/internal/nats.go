package document

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type SnapshotMessage struct {
	DocumentID string `json:"document_id"`
	Content    string `json:"content"`
	Payload    string `json:"payload"`
}

func StartSnapshotConsumer(ctx context.Context, store *Store, natsURL string) (*nats.Conn, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	subject := "doclet.documents.*.snapshots"
	_, err = nc.Subscribe(subject, func(msg *nats.Msg) {
		var payload SnapshotMessage
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("nats snapshot decode error: %v", err)
			return
		}
		docID, err := uuid.Parse(payload.DocumentID)
		if err != nil {
			log.Printf("nats snapshot invalid document_id: %v", err)
			return
		}
		encoded := payload.Content
		if encoded == "" {
			encoded = payload.Payload
		}
		content, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			log.Printf("nats snapshot invalid content: %v", err)
			return
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := store.UpdateContent(ctx, docID, content); err != nil {
			if IsNotFound(err) {
				log.Printf("nats snapshot ignored missing document: %s", docID)
				return
			}
			log.Printf("nats snapshot update error: %v", err)
		}
	})
	if err != nil {
		nc.Close()
		return nil, err
	}

	return nc, nil
}
