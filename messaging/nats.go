package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/globusdigital/feature-toggles/toggle"
	"github.com/nats-io/nats.go"
)

const natsSubject = "feature-toggles"

type Nats struct {
	conn *nats.Conn
}

func NewNats(url string) (Nats, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return Nats{}, fmt.Errorf("connecting to NATS", err)
	}

	return Nats{conn}, nil
}

func (b Nats) Clone() {
	b.conn.Close()
}

func (b Nats) Send(ctx context.Context, flags []toggle.Flag) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	data, err := json.Marshal(flags)
	if err != nil {
		return fmt.Errorf("encoding flags: %v", err)
	}
	if err := b.conn.Publish(natsSubject, data); err != nil {
		return fmt.Errorf("publishing flags: %v", err)
	}

	return nil
}
