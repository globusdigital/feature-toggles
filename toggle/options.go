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
}

func (o getOptions) Apply(opts []Option) getOptions {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

func ForInt(name string, value int64) Option {
	return func(o *getOptions) {
		o.values = append(o.values, ConditionValue{Name: name, Value: value, Type: IntType})
	}
}

func ForFloat(name string, value float64) Option {
	return func(o *getOptions) {
		o.values = append(o.values, ConditionValue{Name: name, Value: value, Type: FloatType})
	}
}

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

func For(values ...ConditionValue) ClientOption {
	for i := range values {
		switch v := values[i].Value.(type) {
		case int64:
			values[i].Type = IntType
		case float64:
			values[i].Type = FloatType
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

func WithEventBus(bus EventBus) ClientOption {
	return func(o *clientOptions) {
		o.eventBus = bus
	}
}

func WithPollingUpdateDuration(d time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.updateDuration = d
	}
}

func WithHttpClient(c *http.Client) ClientOption {
	return func(o *clientOptions) {
		o.httpClient = c
	}
}

func WithLogger(l logger) ClientOption {
	return func(o *clientOptions) {
		o.log = l
	}
}
