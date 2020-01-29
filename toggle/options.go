package toggle

type Option func(o *getOptions)

var (
	Global Option = func(o *getOptions) {
		o.Global = true
	}
)

type getOptions struct {
	Global bool
}

func (o getOptions) Apply(opts []Option) getOptions {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}
