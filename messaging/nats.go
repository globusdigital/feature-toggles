package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	DefaultNatsReconnectBufSize = 16 * 1024 * 1024
	DefaultNatsPingInterval     = time.Minute
	natsSubject                 = "feature-toggles"
)

type Nats struct {
	conn *nats.EncodedConn
}

func NewNats(url string, opts ...nats.Option) (Nats, error) {
	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return Nats{}, fmt.Errorf("connecting to NATS: %w", err)
	}

	encoded, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return Nats{}, fmt.Errorf("creating json encoded connection: %v", err)
	}

	return Nats{encoded}, nil
}

func (b Nats) Close() {
	b.conn.Close()
}

func (b Nats) Send(ctx context.Context, event Event) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := b.conn.Publish(natsSubject, event); err != nil {
		return fmt.Errorf("publishing event: %v", err)
	}

	return nil
}
