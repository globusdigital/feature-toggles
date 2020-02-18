package messaging_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/globusdigital/feature-toggles/messaging"
	"github.com/globusdigital/feature-toggles/toggle"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

var natsURL = "nats://localhost:4222/"

func TestNats_Send(t *testing.T) {
	type args struct {
		ctx   context.Context
		event messaging.Event
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "canceled context", args: args{ctx: canceledCtx()}, wantErr: true},
		{name: "event", args: args{ctx: context.Background(), event: messaging.Event{
			Type: messaging.SaveEvent,
			Flags: []toggle.Flag{{
				Name: "name", ServiceName: "svc1", RawValue: "t", Value: true, Condition: toggle.Condition{
					Op: toggle.OrOp,
					Fields: []toggle.ConditionField{{ConditionValue: toggle.ConditionValue{
						Name:  "userID",
						Type:  toggle.IntType,
						Value: int64(123456),
					}}},
				},
			}},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := messaging.NewNats(natsURL)
			if errors.Is(err, nats.ErrNoServers) {
				t.Skipf("NATs connection error %q, skipping", err)
			}
			a := assert.New(t)
			a.NoError(err)

			var ch chan *nats.Msg
			if !tt.wantErr {
				conn, err := nats.Connect(natsURL)
				a.NoError(err)

				ch = make(chan *nats.Msg)
				sub, err := conn.ChanSubscribe("feature-toggles", ch)
				a.NoError(err)
				defer sub.Unsubscribe()
			}

			if err := b.Send(tt.args.ctx, tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("Nats.Send() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			msg := <-ch
			var ev messaging.Event

			err = json.Unmarshal(msg.Data, &ev)
			a.NoError(err)

			a.Equal(tt.args.event, ev)
		})
	}
}

func canceledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}
