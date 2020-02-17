package toggle_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/globusdigital/feature-toggles/toggle"
)

func TestGet(t *testing.T) {
	type args struct {
		name string
		opts []toggle.Option
	}
	tests := []struct {
		name    string
		cname   string
		seed    []string
		args    args
		sub     bool
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
			for _, s := range tt.seed {
				parts := strings.SplitN(s, "=", 2)
				os.Setenv(parts[0], parts[1])
			}
			toggle.Initialize(context.Background(), tt.cname)

			if tt.sub {
				toggle.DefaultClient = toggle.New(tt.cname)
			}
			if got := toggle.Get(tt.args.name, tt.args.opts...); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
			if got := toggle.GetRaw(tt.args.name, tt.args.opts...); got != tt.wantRaw {
				t.Errorf("Client.Get() = %v, want %v", got, tt.wantRaw)
			}
		})
	}
}
