package toggle_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/globusdigital/feature-toggles/toggle"
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
		ctx       func() context.Context
		enable    bool
		seed      []string
		serverErr bool
		jsonErr   bool
		pollErr   bool
		wantErr   bool
		want      []toggle.Flag
	}{
		{name: "disabled", cname: "serv1", seed: seed1, ctx: canceledCtx(time.Millisecond * 100), want: seedData},
		{name: "canceled ctx", cname: "serv1", ctx: canceledCtx(0), seed: seed1, enable: true},
		{name: "server error", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, serverErr: true, enable: true, wantErr: true},
		{name: "invalid json", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, jsonErr: true, enable: true, wantErr: true},
		{name: "poll json", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, pollErr: true, enable: true, wantErr: true},
		{name: "poll", cname: "serv1", ctx: canceledCtx(time.Second), seed: seed1, enable: true, want: update1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverErr {
					http.Error(w, "error", 500)
					return
				}

				if strings.HasSuffix(r.URL.Path, "/initial") {
					if tt.jsonErr {
						w.Write([]byte(`[{foo:1]`))
						return
					}
					b, err := json.Marshal(initialData)
					a.NoError(err)
					w.Write(b)
				} else {
					if tt.pollErr {
						w.Write([]byte(`[{foo:1]`))
						return
					}
					b, err := json.Marshal(update1)
					a.NoError(err)
					w.Write(b)
				}
			}))
			defer ts.Close()

			if tt.enable {
				tt.seed = append(tt.seed, "FEATURE__GLOBAL__"+toggle.ServerAddressFlag+"="+ts.URL)
			}

			c := toggle.New(tt.cname)
			c.ParseEnv(tt.seed)

			toggle.UpdateDuration = 100 * time.Millisecond
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
						value = c.GetRaw(f.Name, toggle.Global)
					} else if f.ServiceName == tt.cname {
						value = c.GetRaw(f.Name)
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
