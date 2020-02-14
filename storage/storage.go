package storage

import (
	"context"

	"github.com/globusdigital/feature-toggles/toggle"
)

type Kind int

const (
	MemKind   Kind = iota // mem
	MongoKind             // mongo
)

type Store interface {
	Get(ctx context.Context, serviceName string) ([]toggle.Flag, error)
	Save(ctx context.Context, flags []toggle.Flag, initial bool) error
	Delete(ctx context.Context, flags []toggle.Flag) error
}
