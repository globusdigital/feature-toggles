package toggle

import (
	"context"
	"os"
)

var (
	// ServerAddressFlag is a feature flag which enables working with a centralized
	// toggle server. It's value is the server address
	ServerAddressFlag = "TOGGLE_SERVER"

	// DefaultClient is the client that is used for the global access functions
	DefaultClient *Client
)

// Initialize creates the global DefaultClient instance, parses the local
// environment and attempts to connect to a feature toggles server.
func Initialize(ctx context.Context, name string, opts ...ClientOption) {
	DefaultClient = New(name, opts...)
	DefaultClient.ParseEnv(os.Environ())

	go func() {
		for err := range DefaultClient.Connect(ctx) {
			DefaultClient.opts.log.Println("Error listening for updates:", err)
		}
	}()
}

// Get calls the DefaultClient.Get method
func Get(name string, opts ...Option) bool {
	return DefaultClient.Get(name, opts...)
}

// GetRaw calls the DefaultClient.GetRaw method
func GetRaw(name string, opts ...Option) string {
	return DefaultClient.GetRaw(name, opts...)
}
