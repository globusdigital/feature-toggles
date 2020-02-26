package toggle

import "context"

type EventType string

const (
	SaveEvent   EventType = "save"
	DeleteEvent EventType = "delete"
	ErrorEvent  EventType = "error"
)

type Event struct {
	Type  EventType `json:"type"`
	Flags []Flag    `json:"flags"`
	Error string    `json:"error"`
}

type EventBus interface {
	Receiver(ctx context.Context) <-chan Event
}
