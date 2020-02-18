package toggle

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	// ServerAddressFlag is a feature flag which enables working with a centralized
	// toggle server. It's value is the server address
	ServerAddressFlag = "TOGGLE_SERVER"

	// UpdateDuration is the update cycle duration when the toggle server is implemented
	UpdateDuration = 30 * time.Minute

	// HttpClient is the http client used to communicate with the server
	HttpClient = http.DefaultClient

	// DefaultClient is the client that is used for the global access functions
	DefaultClient *Client
)

// Initialize creates the global DefaultClient instance, parses the local
// environment and attempts to connect to a feature toggles server.
func Initialize(ctx context.Context, name string) {
	DefaultClient = New(name)
	DefaultClient.ParseEnv(os.Environ())

	go func() {
		c := DefaultClient.Connect(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-c:
				log.Println("Error listening for updates:", err)
			}
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
