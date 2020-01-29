package toggle

import (
	"os"
)

var DefaultClient *Client

func Initialize(name string) {
	DefaultClient = New(name)
	DefaultClient.ParseEnv(os.Environ())
}

func Get(name string, opts ...Option) bool {
	return DefaultClient.Get(name, opts...)
}
