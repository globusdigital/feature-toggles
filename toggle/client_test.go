package toggle_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/globusdigital/feature-toggles/toggle"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	seed1 = []string{
		"SHELL=/bin/zsh",
		"FEATURE_SERV1_FEATURE_1=t",
		"FEATURE_SERV1_FEATURE_2=f",
		"FEATURE_SERV1_FEATURE_3=yes",
		"FEATURE_SERV1_FEATURE_4=1",
		"FEATURE_SERV2_FEATURE_5=y",
		"FEATURE_SERV2_FEATURE_6=true",
		"FEATURE__GLOBAL__SOME_SHARED_FEATURE=y",
	}

	seedData = []toggle.Flag{
		{Name: "feature.1", ServiceName: "serv1", RawValue: "t", Value: true},
		{Name: "feature.2", ServiceName: "serv1", RawValue: "f"},
		{Name: "feature.3", ServiceName: "serv1", RawValue: "yes", Value: true},
		{Name: "feature.4", ServiceName: "serv1", RawValue: "1", Value: true},
		{Name: "feature.5", ServiceName: "serv2", RawValue: "y", Value: true},
		{Name: "feature.6", ServiceName: "serv2", RawValue: "true", Value: true},
		{Name: "some.shared.feature", ServiceName: "", RawValue: "y", Value: true},
	}
	initialData = []toggle.Flag{
		{Name: "feature.1", ServiceName: "serv1", RawValue: "t", Value: true},
		{Name: "feature.2", ServiceName: "serv1", RawValue: "0"},
		{Name: "feature.3", ServiceName: "serv1", RawValue: "no"},
		{Name: "feature.5", ServiceName: "serv2", RawValue: "1", Value: true},
		{Name: "some.shared.feature", ServiceName: "", RawValue: "t", Value: true},
	}
	update1 = []toggle.Flag{
		{Name: "feature.1", ServiceName: "serv1", RawValue: "1", Value: true},
		{Name: "feature.2", ServiceName: "serv1", RawValue: "t", Value: true},
		{Name: "feature.3", ServiceName: "serv1", RawValue: "yes", Value: true},
		{Name: "feature.5", ServiceName: "serv2", RawValue: "1", Value: true},
		{Name: "feature.6", ServiceName: "serv2", RawValue: "some data"},
		{Name: "some.shared.feature", ServiceName: "", RawValue: "t", Value: true},
	}

	cond1 = []toggle.Flag{
		{Name: "feature.1", ServiceName: "serv1", RawValue: "1", Value: true},
		{Name: "some.shared.feature", ServiceName: "", RawValue: "t", Value: true, Condition: toggle.Condition{
			Op: toggle.OrOp,
			Fields: []toggle.ConditionField{
				{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue, Type: toggle.StringType, Value: "serv1"}},
				{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue, Type: toggle.StringType, Value: "serv3"}},
			},
		}},
	}

	cond2 = []toggle.Flag{
		{Name: "feature.1", ServiceName: "serv1", RawValue: "some value", Condition: toggle.Condition{
			Fields: []toggle.ConditionField{
				{ConditionValue: toggle.ConditionValue{Name: "userID", Type: toggle.IntType, Value: int64(10)}, Op: toggle.LtOp},
			},
		}},
	}

	ev1Data = []toggle.Flag{
		{Name: "feature.2", ServiceName: "serv1", RawValue: "0"},
		{Name: "some.shared.feature", ServiceName: "", RawValue: "t", Value: true},
		{Name: "feature.1", ServiceName: "serv2", RawValue: "t", Value: true},
		{Name: "feature.3", ServiceName: "serv1", RawValue: "1", Value: true},
		{Name: "feature.4", ServiceName: "serv2", RawValue: "some data"},
	}

	ev2Data = []toggle.Flag{
		{Name: "feature.2", ServiceName: "serv1", RawValue: "0"},
		{Name: "some.shared.feature", ServiceName: "", RawValue: "t", Value: true},
		{Name: "feature.1", ServiceName: "serv2", RawValue: "t", Value: true},
		{Name: "feature.3", ServiceName: "serv1", RawValue: "n"},
		{Name: "feature.4", ServiceName: "serv2", RawValue: "some data"},
	}
)

func TestClient_Get(t *testing.T) {
	type args struct {
		name string
		opts []toggle.Option
	}
	tests := []struct {
		name    string
		cname   string
		seed    []string
		args    args
		want    bool
		wantRaw string
	}{
		{name: "serv1 feat1", cname: "serv1", seed: seed1, args: args{name: "feature.1"}, want: true, wantRaw: "t"},
		{name: "serv1 feat2", cname: "serv1", seed: seed1, args: args{name: "feature.2"}, wantRaw: "f"},
		{name: "serv1 feat3", cname: "serv1", seed: seed1, args: args{name: "feature.3"}, want: true, wantRaw: "yes"},
		{name: "serv1 feat4", cname: "serv1", seed: seed1, args: args{name: "feature.4"}, want: true, wantRaw: "1"},
		{name: "serv1 feat5", cname: "serv1", seed: seed1, args: args{name: "feature.5"}},
		{name: "serv1 feat6", cname: "serv1", seed: seed1, args: args{name: "feature.6"}},
		{name: "serv2 feat4", cname: "serv2", seed: seed1, args: args{name: "feature.4"}},
		{name: "serv2 feat5", cname: "serv2", seed: seed1, args: args{name: "feature.5"}, want: true, wantRaw: "y"},
		{name: "serv2 feat6", cname: "serv2", seed: seed1, args: args{name: "feature.6"}, want: true, wantRaw: "true"},
		{name: "global feat6", cname: "serv1", seed: seed1, args: args{name: "feature.6", opts: []toggle.Option{toggle.Global}}},
		{name: "global feat1", cname: "serv2", seed: seed1, args: args{name: "feature.1", opts: []toggle.Option{toggle.Global}}},
		{name: "global some shared feature", cname: "serv1", seed: seed1, args: args{name: "some.shared.feature", opts: []toggle.Option{toggle.Global}}, want: true, wantRaw: "y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := toggle.New(tt.cname)
			c.ParseEnv(tt.seed)

			if got := c.Get(tt.args.name, tt.args.opts...); got != tt.want {
				t.Errorf("Client.Get() = %v, want %v", got, tt.want)
			}

			if got := c.GetRaw(tt.args.name, tt.args.opts...); got != tt.wantRaw {
				t.Errorf("Client.Get() = %v, want %v", got, tt.wantRaw)
			}
		})
	}
}

func TestClient_Connect(t *testing.T) {
	tests := []struct {
		name      string
		cname     string
		copts     []toggle.ClientOption
		ctx       func() context.Context
		enable    bool
		seed      []string
		serverErr bool
		jsonErr   bool
		pollErr   bool
		apiPath   string
		ev        []toggle.Event
		evErr     bool
		update    []toggle.Flag
		opts      []toggle.Option
		wantErr   bool
		want      []toggle.Flag
	}{
		{name: "disabled", cname: "serv1", seed: seed1, ctx: canceledCtx(time.Millisecond * 100), want: seedData},
		{name: "canceled ctx", cname: "serv1", ctx: canceledCtx(0), seed: seed1, enable: true},
		{name: "server error", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, serverErr: true, enable: true, wantErr: true},
		{name: "invalid json", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, jsonErr: true, enable: true, wantErr: true},
		{name: "poll json", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, pollErr: true, enable: true, wantErr: true},
		{name: "poll", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, enable: true, update: update1, want: update1},
		{name: "conditional 1", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, enable: true, update: cond1, want: cond1},
		{name: "conditional 1 - serv2", cname: "serv2", ctx: canceledCtx(time.Second), seed: seed1, enable: true, update: cond1},
		{name: "conditional 2", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, enable: true, update: cond2, want: []toggle.Flag{{Name: "feature.1", ServiceName: "serv1"}}},
		{name: "conditional 2 - val 20", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, enable: true, update: cond2, opts: []toggle.Option{toggle.ForInt("userID", 20)}, want: cond2},
		{name: "event err", cname: "serv1", ctx: canceledCtx(50 * time.Millisecond), seed: seed1, enable: true, evErr: true, want: initialData},
		{name: "event 1", cname: "serv1", ctx: canceledCtx(50 * time.Millisecond), seed: seed1, enable: true, ev: []toggle.Event{
			{Type: toggle.SaveEvent, Flags: []toggle.Flag{
				{Name: "feature.1", ServiceName: "serv2", RawValue: "t", Value: true},
			}},
			{Type: toggle.DeleteEvent, Flags: initialData[0:1]},
			{Type: toggle.DeleteEvent, Flags: initialData[2:4]},
			{Type: toggle.SaveEvent, Flags: ev1Data[3:]},
		}, want: ev1Data},
		{name: "event 1 - path 2", cname: "serv1", ctx: canceledCtx(50 * time.Millisecond), seed: seed1, enable: true, ev: []toggle.Event{
			{Type: toggle.SaveEvent, Flags: []toggle.Flag{
				{Name: "feature.1", ServiceName: "serv2", RawValue: "t", Value: true},
			}},
			{Type: toggle.DeleteEvent, Flags: initialData[0:1]},
			{Type: toggle.DeleteEvent, Flags: initialData[2:4]},
			{Type: toggle.SaveEvent, Flags: ev1Data[3:]},
			{Type: toggle.SaveEvent, Flags: ev2Data[3:]},
		}, apiPath: "/service/featuretoggles", copts: []toggle.ClientOption{toggle.WithAPIPath("/service/featuretoggles")}, want: ev2Data},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverErr {
					http.Error(w, "error", 500)
					return
				}

				if tt.apiPath == "" {
					tt.apiPath = "/flags"
				}
				if !strings.Contains(r.URL.Path, tt.apiPath) {
					t.Fatalf("Invalid request path %s", r.URL.Path)
				}

				if strings.HasSuffix(r.URL.Path, "/initial") {
					if tt.jsonErr {
						_, _ = w.Write([]byte(`[{foo:1]`))
						return
					}
					b, err := json.Marshal(initialData)
					a.NoError(err)
					_, _ = w.Write(b)
				} else {
					if tt.pollErr {
						_, _ = w.Write([]byte(`[{foo:1]`))
						return
					}
					b, err := json.Marshal(tt.update)
					a.NoError(err)
					_, _ = w.Write(b)
				}
			}))
			defer ts.Close()

			if tt.enable {
				tt.seed = append(tt.seed, "FEATURE__GLOBAL__"+toggle.ServerAddressFlag+"="+ts.URL)
			}

			copts := tt.copts
			if len(tt.ev) > 0 {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				bus := NewMockEventBus(ctrl)

				var ch chan toggle.Event
				if len(tt.ev) > 0 {
					ch = make(chan toggle.Event)

					go func() {
						defer close(ch)
						for _, ev := range tt.ev {
							ch <- ev
						}
					}()
				}
				var err error
				if tt.evErr {
					err = errors.New("ev")
				}
				bus.EXPECT().Receive(gomock.Any()).AnyTimes().Return(ch, err)

				copts = append(copts, toggle.WithEventBus(bus))
			}
			if len(copts) == 0 && !tt.evErr {
				copts = append(copts, toggle.WithPollingUpdateDuration(100*time.Millisecond))
			}
			c := toggle.New(tt.cname, copts...)
			c.ParseEnv(tt.seed)

			ctx := tt.ctx()
			got := c.Connect(ctx)

			if tt.wantErr {
				select {
				case <-ctx.Done():
					t.Errorf("Expected error")
				case err := <-got:
					a.Error(err)
				}
			} else {
				<-ctx.Done()

				for _, f := range tt.want {
					var value string
					if f.ServiceName == "" {
						opts := append([]toggle.Option{toggle.Global}, tt.opts...)
						value = c.GetRaw(f.Name, opts...)
					} else if f.ServiceName == tt.cname {
						value = c.GetRaw(f.Name, tt.opts...)
					} else {
						continue
					}

					a.Equal(f.RawValue, value, f.Name)
				}

			}
		})
	}
}

func canceledCtx(d time.Duration) func() context.Context {
	return func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(d, cancel)

		return ctx
	}
}
