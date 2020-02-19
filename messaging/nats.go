package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/globusdigital/feature-toggles/toggle"
	"github.com/nats-io/nats.go"
)

const (
	DefaultNatsReconnectBufSize = 16 * 1024 * 1024
	DefaultNatsPingInterval     = time.Minute
	NatsSubject                 = "feature-toggles"
)

type Nats struct {
	Conn *nats.EncodedConn
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
	b.Conn.Close()
}

func (b Nats) Send(ctx context.Context, event toggle.Event) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := b.Conn.Publish(NatsSubject, event); err != nil {
		return fmt.Errorf("publishing event: %v", err)
	}

	return nil
}

func (s Nats) Receive(ctx context.Context) (<-chan toggle.Event, error) {
	ch := make(chan toggle.Event)
	sub, err := s.Conn.Subscribe(NatsSubject, func(ev toggle.Event) {
		ch <- ev
	})

	if err != nil {
		return nil, fmt.Errorf("subscribing to nats subject: %v", err)
	}

	go func() {
		defer close(ch)
		defer sub.Unsubscribe()
		select {
		case <-ctx.Done():
			return
		}
	}()

	return ch, nil
}
