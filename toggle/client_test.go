package toggle_test

import (
	"testing"

	"github.com/globusdigital/feature-toggles/toggle"
)

func TestClient_Get(t *testing.T) {
	type args struct {
		name string
		opts []toggle.Option
	}
	tests := []struct {
		name  string
		cname string
		seed  []string
		args  args
		want  bool
	}{
		{name: "serv1 feat1", cname: "serv1", seed: seed1, args: args{name: "feature.1"}, want: true},
		{name: "serv1 feat2", cname: "serv1", seed: seed1, args: args{name: "feature.2"}},
		{name: "serv1 feat3", cname: "serv1", seed: seed1, args: args{name: "feature.3"}, want: true},
		{name: "serv1 feat4", cname: "serv1", seed: seed1, args: args{name: "feature.4"}, want: true},
		{name: "serv1 feat5", cname: "serv1", seed: seed1, args: args{name: "feature.5"}},
		{name: "serv1 feat6", cname: "serv1", seed: seed1, args: args{name: "feature.6"}},
		{name: "serv2 feat4", cname: "serv2", seed: seed1, args: args{name: "feature.4"}},
		{name: "serv2 feat5", cname: "serv2", seed: seed1, args: args{name: "feature.5"}, want: true},
		{name: "serv2 feat6", cname: "serv2", seed: seed1, args: args{name: "feature.6"}, want: true},
		{name: "global feat6", cname: "serv1", seed: seed1, args: args{name: "feature.6", opts: []toggle.Option{toggle.Global}}},
		{name: "global feat1", cname: "serv2", seed: seed1, args: args{name: "feature.1", opts: []toggle.Option{toggle.Global}}},
		{name: "global some shared feature", cname: "serv1", seed: seed1, args: args{name: "some.shared.feature", opts: []toggle.Option{toggle.Global}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := toggle.New(tt.cname)
			c.ParseEnv(tt.seed)

			if got := c.Get(tt.args.name, tt.args.opts...); got != tt.want {
				t.Errorf("Client.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
)
