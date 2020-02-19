package toggle

import "context"

type EventType string

const (
	SaveEvent   EventType = "save"
	DeleteEvent EventType = "delete"
)

type Event struct {
	Type  EventType `json:"type"`
	Flags []Flag    `json:"flags"`
}

type EventBus interface {
	Receive(ctx context.Context) (<-chan Event, error)
}
