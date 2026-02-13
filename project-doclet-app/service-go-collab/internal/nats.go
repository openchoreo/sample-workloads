package collab

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/nats-io/nats.go"
)

type NatsBroker struct {
	nc *nats.Conn
}

func NewNatsBroker(url string) (*NatsBroker, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NatsBroker{nc: nc}, nil
}

func (b *NatsBroker) Close() {
	if b.nc != nil {
		b.nc.Close()
	}
}

func (b *NatsBroker) Publish(subject string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("nats marshal error: %v", err)
		return
	}
	if err := b.nc.Publish(subject, data); err != nil {
		log.Printf("nats publish error: %v", err)
	}
}

func (b *NatsBroker) Subscribe(subject string, handler func(Message)) (*nats.Subscription, error) {
	return b.nc.Subscribe(subject, func(msg *nats.Msg) {
		var payload Message
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("nats decode error: %v", err)
			return
		}
		handler(payload)
	})
}

func SubjectForDocument(docID, suffix string) string {
	clean := strings.TrimSpace(docID)
	if clean == "" {
		clean = "unknown"
	}
	return "doclet.documents." + clean + "." + suffix
}
