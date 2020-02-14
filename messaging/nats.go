package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

const natsSubject = "feature-toggles"

type Nats struct {
	conn *nats.Conn
}

func NewNats(url string) (Nats, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return Nats{}, fmt.Errorf("connecting to NATS: %v", err)
	}

	return Nats{conn}, nil
}

func (b Nats) Clone() {
	b.conn.Close()
}

func (b Nats) Send(ctx context.Context, event Event) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encoding evebt: %v", err)
	}
	if err := b.conn.Publish(natsSubject, data); err != nil {
		return fmt.Errorf("publishing event: %v", err)
	}

	return nil
}
