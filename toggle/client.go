package toggle

import (
	"strings"
	"sync"
	"unicode"
)

type Client struct {
	name  string
	store map[string][]Flag
	mu    sync.RWMutex
}

type Flag struct {
	ServiceName string

	RawValue string
	Value    bool
}

func New(name string) *Client {
	return &Client{name: name, store: map[string][]Flag{}}
}

func (c *Client) Get(name string, opts ...Option) bool {
	o := (getOptions{}).Apply(opts)

	f := c.getFlag(name, o)

	return f.Value
}

func (c *Client) getFlag(name string, o getOptions) Flag {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, f := range c.store[name] {
		if f.ServiceName != c.name && (f.ServiceName != "" || !o.Global) {
			continue
		}

		return f
	}

	return Flag{}
}

const (
	featurePrefix = "FEATURE_"
	globalName    = "_GLOBAL__"
)

func (c *Client) ParseEnv(env []string) {
	flags := map[string][]Flag{}

	for _, e := range env {
		if !strings.HasPrefix(e, featurePrefix) {
			continue
		}
		eqIdx := strings.IndexByte(e, '=')
		rawValue := e[eqIdx+1:]
		if rawValue == "" {
			continue
		}

		key := e[len(featurePrefix):eqIdx]

		var serviceName string
		if strings.HasPrefix(key, globalName) {
			key = key[len(globalName):]
		} else {
			underIdx := strings.IndexByte(key, '_')
			serviceName = key[:underIdx]
			key = key[underIdx+1:]
		}

		var value bool
		switch strings.ToLower(rawValue) {
		case "1", "y", "yes", "t", "true":
			value = true
		}

		key = normalizeKey(key)

		flags[key] = append(flags[key], Flag{
			ServiceName: normalizeSerivceName(serviceName),
			RawValue:    rawValue,
			Value:       value,
		})
	}

	c.mu.Lock()
	c.store = flags
	c.mu.Unlock()
}

func normalizeSerivceName(name string) string {
	name = strings.ToLower(name)

	return name
}

func normalizeKey(value string) string {
	return strings.Map(func(r rune) rune {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
			return '.'
		}
		return unicode.ToLower(r)
	}, value)
}
