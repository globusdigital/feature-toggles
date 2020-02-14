package messaging

import (
	"context"

	"github.com/globusdigital/feature-toggles/toggle"
)

type Kind int
type EventType string

const (
	NoopKind Kind = iota // noop
	NatsKind             // nats

	SaveEvent   EventType = "save"
	DeleteEvent EventType = "delete"
)

type Event struct {
	Type  EventType     `json:"type"`
	Flags []toggle.Flag `json:"flags"`
}

type Bus interface {
	Send(ctx context.Context, event Event) error
}

type Noop struct{}

func NewNoop() Noop {
	return Noop{}
}

func (s Noop) Send(ctx context.Context, event Event) error {
	return ctx.Err()
}
