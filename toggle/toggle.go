package toggle

import (
	"context"
	"os"
)

var DefaultClient *Client

func Initialize(ctx context.Context, name string) {
	DefaultClient = New(name)
	DefaultClient.ParseEnv(os.Environ())
	DefaultClient.Connect(ctx)
}

func Get(name string, opts ...Option) bool {
	return DefaultClient.Get(name, opts...)
}

func GetRaw(name string, opts ...Option) string {
	return DefaultClient.GetRaw(name, opts...)
}
