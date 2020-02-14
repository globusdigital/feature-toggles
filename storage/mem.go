package storage

import (
	"context"
	"sync"

	"github.com/globusdigital/feature-toggles/toggle"
)

type flagKey struct {
	name, serviceName string
}

type Mem struct {
	data map[flagKey]toggle.Flag
	mu   sync.RWMutex
}

func NewMem() *Mem {
	return &Mem{data: map[flagKey]toggle.Flag{}}
}

func (s *Mem) Get(ctx context.Context, serviceName string) ([]toggle.Flag, error) {
	ret := make([]toggle.Flag, 0, len(s.data))

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, f := range s.data {
		if f.ServiceName == "" || f.ServiceName == serviceName || serviceName == "" {
			ret = append(ret, f)
		}
	}
	return ret, nil
}

func (s *Mem) Save(ctx context.Context, flags []toggle.Flag, initial bool) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, f := range flags {
		key := flagKey{f.Name, f.ServiceName}
		if _, ok := s.data[key]; !initial || !ok {
			s.data[key] = f
		}
	}
	return nil
}
