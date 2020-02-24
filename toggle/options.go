package toggle

import (
	"fmt"
	"net/http"
	"time"
)

type Option func(o *getOptions)
type ClientOption func(o *clientOptions)

var (
	// Global indicates that the flag lookup should search for global flags.
	Global Option = func(o *getOptions) {
		o.global = true
	}
)

type logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

type getOptions struct {
	global bool
	values []ConditionValue
}

type clientOptions struct {
	values         []ConditionValue
	eventBus       EventBus
	updateDuration time.Duration
	httpClient     *http.Client
	log            logger
	path           string
}

func (o getOptions) Apply(opts []Option) getOptions {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// ForInt sets an integer value when querying a flag constraint
func ForInt(name string, value int64) Option {
	return func(o *getOptions) {
		o.values = append(o.values, ConditionValue{Name: name, Value: value, Type: IntType})
	}
}

// ForFloat sets a float value when querying a flag constraint
func ForFloat(name string, value float64) Option {
	return func(o *getOptions) {
		o.values = append(o.values, ConditionValue{Name: name, Value: value, Type: FloatType})
	}
}

// ForBool sets a boolean value when querying a flag constraint
func ForBool(name string, value bool) Option {
	return func(o *getOptions) {
		o.values = append(o.values, ConditionValue{Name: name, Value: value, Type: BoolType})
	}
}

// ForString sets a string value when querying a flag constraint
func ForString(name string, value string) Option {
	return func(o *getOptions) {
		o.values = append(o.values, ConditionValue{Name: name, Value: value, Type: StringType})
	}
}

func (o clientOptions) Apply(opts []ClientOption) clientOptions {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// For sets the global condition values that can be used for any flag
// constraints. The service name is an always present global constraint value.
func For(values ...ConditionValue) ClientOption {
	for i := range values {
		switch v := values[i].Value.(type) {
		case int64:
			values[i].Type = IntType
		case float64:
			values[i].Type = FloatType
		case bool:
			values[i].Type = BoolType
		case string:
			values[i].Type = StringType
		default:
			panic(fmt.Sprintf("Unsupported type: %T", v))
		}
	}

	return func(o *clientOptions) {
		o.values = append(o.values, values...)
	}
}

// WithEventBus sets the event bus to be used for listening for events on flag updates
func WithEventBus(bus EventBus) ClientOption {
	return func(o *clientOptions) {
		o.eventBus = bus
	}
}

// WithPollingUpdateDuration sets the duration between poll iterations.
// Defaults to 30m
func WithPollingUpdateDuration(d time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.updateDuration = d
	}
}

// WithHttpClient sets the http client used for contacting the flags server.
// Defaults to http.DefaultClient
func WithHttpClient(c *http.Client) ClientOption {
	return func(o *clientOptions) {
		o.httpClient = c
	}
}

// WithLogger sets the logger object to be used when errors and information are
// logged. Defaults to the standard logger to STDERR
func WithLogger(l logger) ClientOption {
	return func(o *clientOptions) {
		o.log = l
	}
}

// WithAPIPath sets the API endpoint path for the flags server. Defaults to
// '/flags'
func WithAPIPath(p string) ClientOption {
	return func(o *clientOptions) {
		o.path = p
	}
}
