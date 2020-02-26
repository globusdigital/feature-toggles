package messaging

import (
	"context"

	"github.com/globusdigital/feature-toggles/toggle"
)

type Kind int

const (
	NoopKind Kind = iota // noop
	NatsKind             // nats
)

type Noop struct{}

func NewNoop() Noop {
	return Noop{}
}

func (s Noop) Send(ctx context.Context, event toggle.Event) error {
	return ctx.Err()
}

func (s Noop) Receiver(ctx context.Context) <-chan toggle.Event {
	return nil
}
