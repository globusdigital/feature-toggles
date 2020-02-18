package toggle

import (
	"fmt"
)

type Option func(o *getOptions)
type ClientOption func(o *clientOptions)

var (
	// Global indicates that the flag lookup should search for global flags.
	Global Option = func(o *getOptions) {
		o.Global = true
	}
)

type getOptions struct {
	Global bool
	Values []ConditionValue
}

type clientOptions struct {
	Values []ConditionValue
}

func (o getOptions) Apply(opts []Option) getOptions {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

func ForInt(name string, value int64) Option {
	return func(o *getOptions) {
		o.Values = append(o.Values, ConditionValue{Name: name, Value: value, Type: IntValue})
	}
}

func ForFloat(name string, value float64) Option {
	return func(o *getOptions) {
		o.Values = append(o.Values, ConditionValue{Name: name, Value: value, Type: FloatValue})
	}
}

func ForString(name string, value string) Option {
	return func(o *getOptions) {
		o.Values = append(o.Values, ConditionValue{Name: name, Value: value, Type: StringValue})
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
			values[i].Type = IntValue
		case float64:
			values[i].Type = FloatValue
		case string:
			values[i].Type = StringValue
		default:
			panic(fmt.Sprintf("Unsupported type: %T", v))
		}
	}

	return func(o *clientOptions) {
		o.Values = append(o.Values, values...)
	}
}
