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

type Bus interface {
	Send(ctx context.Context, flags []toggle.Flag) error
}

type Noop struct{}

func NewNoop() Noop {
	return Noop{}
}

func (s Noop) Send(ctx context.Context, flags []toggle.Flag) error {
	return ctx.Err()
}
