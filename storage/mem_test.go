package storage

import (
	"context"
	"sort"
	"testing"

	"github.com/globusdigital/feature-toggles/toggle"
	"github.com/stretchr/testify/assert"
)

var initialData = []toggle.Flag{
	{Name: "n1", ServiceName: "svc1", RawValue: "t", Value: true},
	{Name: "n2", ServiceName: "svc1", RawValue: "0"},
	{Name: "n3", ServiceName: "svc2", RawValue: "1", Value: true},
	{Name: "n4", ServiceName: "", RawValue: "some data"},
	{Name: "n5", ServiceName: "", RawValue: "y", Value: true},
}

func TestMem_Get(t *testing.T) {
	type args struct {
		ctx         context.Context
		serviceName string
	}
	tests := []struct {
		name    string
		args    args
		initial []toggle.Flag
		want    []toggle.Flag
		wantErr bool
	}{
		{name: "canceled ctx", args: args{ctx: canceledCtx()}, wantErr: true},
		{name: "no data", args: args{ctx: context.Background()}},
		{name: "no data - svc", args: args{ctx: context.Background(), serviceName: "svc1"}},
		{name: "data", args: args{ctx: context.Background()}, initial: initialData, want: initialData},
		{name: "data - svc1", args: args{ctx: context.Background(), serviceName: "svc1"}, initial: initialData, want: []toggle.Flag{
			initialData[0], initialData[1], initialData[3], initialData[4],
		}},
		{name: "data - svc2", args: args{ctx: context.Background(), serviceName: "svc2"}, initial: initialData, want: initialData[2:]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			s := NewMem()

			if tt.initial != nil {
				err := s.Save(context.Background(), tt.initial, false)
				a.NoError(err)
			}

			got, err := s.Get(tt.args.ctx, tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mem.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].Name < got[j].Name
			})
			a.Equal(tt.want, got)
		})
	}
}

func TestMem_Save(t *testing.T) {
	type args struct {
		ctx     context.Context
		flags   []toggle.Flag
		initial bool
	}
	tests := []struct {
		name     string
		args     args
		initial  []toggle.Flag
		expected []toggle.Flag
		wantErr  bool
	}{
		{name: "canceled ctx", args: args{ctx: canceledCtx()}, wantErr: true},
		{name: "save data", args: args{ctx: context.Background(), flags: initialData}, expected: initialData},
		{name: "save data after initial", args: args{ctx: context.Background(), flags: []toggle.Flag{
			{Name: "n2", ServiceName: "svc1", RawValue: "1", Value: true},
			{Name: "n3", ServiceName: "svc1", RawValue: "0"},
		}, initial: true}, initial: initialData, expected: []toggle.Flag{
			initialData[0], initialData[1], initialData[2], initialData[3], initialData[4],
			{Name: "n3", ServiceName: "svc1", RawValue: "0"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			s := NewMem()

			if tt.initial != nil {
				err := s.Save(context.Background(), tt.initial, false)
				a.NoError(err)
			}

			err := s.Save(tt.args.ctx, tt.args.flags, tt.args.initial)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mem.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			a.Len(s.data, len(tt.expected))

			for _, v := range s.data {
				var found bool
				for _, e := range tt.expected {
					if e == v {
						found = true
					}
				}
				a.True(found, "expected %#v", v)
			}
		})
	}
}

func canceledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}
