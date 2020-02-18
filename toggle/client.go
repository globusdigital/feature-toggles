package toggle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	featurePrefix = "FEATURE_"
	globalName    = "_GLOBAL__"
)

type Client struct {
	name string
	opts clientOptions

	store map[string][]Flag
	mu    sync.RWMutex
}

type Flag struct {
	Name        string `json:"name,omitempty"`
	ServiceName string `json:"serviceName"`

	RawValue string `json:"rawValue"`
	Value    bool   `json:"value"`

	Condition Condition `json:"condition"`
}

// New creates a new toggle client with the given service name
func New(name string, opts ...ClientOption) *Client {
	o := (clientOptions{Values: []ConditionValue{
		{Name: ServiceNameValue, Type: StringValue, Value: name},
	}}).Apply(opts)

	return &Client{name: normalizeSerivceName(name), opts: o, store: map[string][]Flag{}}
}

// Get returns the boolean flag value
func (c *Client) Get(name string, opts ...Option) bool {
	o := (getOptions{}).Apply(opts)

	f := c.getFlag(name, o)

	return f.Value
}

// GetRaw returns the raw string flag value
func (c *Client) GetRaw(name string, opts ...Option) string {
	o := (getOptions{}).Apply(opts)

	f := c.getFlag(name, o)

	return f.RawValue
}

func (c *Client) getFlag(name string, o getOptions) Flag {
	c.mu.RLock()
	defer c.mu.RUnlock()

	name = normalizeName(name)
	for _, f := range c.store[name] {
		if f.ServiceName != c.name && (f.ServiceName != "" || !o.Global) {
			continue
		}

		values := make([]ConditionValue, len(c.opts.Values)+len(o.Values))
		copy(values, c.opts.Values)
		copy(values[len(c.opts.Values):], o.Values)
		if !f.Condition.Match(values) {
			break
		}

		return f
	}

	return Flag{}
}

// ParseEnv parses the given environment variables and populates the flags
func (c *Client) ParseEnv(env []string) {
	flags := map[string][]Flag{}

	for _, e := range env {
		if !strings.HasPrefix(e, featurePrefix) {
			continue
		}
		eqIdx := strings.IndexByte(e, '=')
		rawValue := e[eqIdx+1:]
		if rawValue == "" || eqIdx == -1 {
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

		key = normalizeName(key)

		flags[key] = append(flags[key], Flag{
			Name:        key,
			ServiceName: normalizeSerivceName(serviceName),
			RawValue:    rawValue,
			Value:       value,
		})
	}

	c.mu.Lock()
	c.store = flags
	c.mu.Unlock()
}

func (c *Client) Connect(ctx context.Context) chan error {
	addr := c.GetRaw(ServerAddressFlag, Global)
	if addr == "" {
		return nil
	}

	errC := make(chan error)

	go func() {
		if err := c.seedFlags(ctx, addr); err != nil {
			errC <- err
			return
		}

		if ctx.Err() != nil {
			return
		}

		ticker := time.NewTicker(UpdateDuration)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := c.pollFlags(ctx, addr); err != nil {
					errC <- err
					return
				}
			}
		}
	}()

	return errC
}

func (c *Client) seedFlags(ctx context.Context, addr string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	c.mu.RLock()
	data := make([]Flag, 0, len(c.store))

	for _, flags := range c.store {
		data = append(data, flags...)
	}
	c.mu.RUnlock()

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encoding initial flag data: %v", err)
	}

	r, err := http.NewRequestWithContext(ctx, "POST", addr+"/"+path.Join("flags", c.name, "initial"), bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("creating initial flag request: %v", err)
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := HttpClient.Do(r)
	if err != nil {
		return fmt.Errorf("getting initial flag response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d (%s)", resp.StatusCode, resp.Status)
	}

	return c.updateStore(resp.Body)
}

func (c *Client) pollFlags(ctx context.Context, addr string) error {
	r, err := http.NewRequestWithContext(ctx, "GET", addr+"/"+path.Join("flags", c.name), nil)
	if err != nil {
		return fmt.Errorf("creating update poll flag request: %v", err)
	}

	resp, err := HttpClient.Do(r)
	if err != nil {
		return fmt.Errorf("getting update poll flag response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d (%s)", resp.StatusCode, resp.Status)
	}

	return c.updateStore(resp.Body)
}

func (c *Client) updateStore(r io.Reader) error {
	var flags []Flag
	if err := json.NewDecoder(r).Decode(&flags); err != nil {
		return fmt.Errorf("decoding flag data: %v", err)
	}

	store := map[string][]Flag{}
	for _, f := range flags {
		f.Name = normalizeName(f.Name)
		f.ServiceName = normalizeSerivceName(f.ServiceName)
		store[f.Name] = append(store[f.Name], f)
	}

	c.mu.Lock()
	c.store = store
	c.mu.Unlock()

	return nil
}

func normalizeSerivceName(name string) string {
	name = strings.ToLower(name)

	return name
}

func normalizeName(value string) string {
	return strings.Map(func(r rune) rune {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
			return '.'
		}
		return unicode.ToLower(r)
	}, value)
}

func (f Flag) String() string {
	if f.Condition.hasMatchers() {
		return fmt.Sprintf("%s=%s %s", f.Name, f.RawValue, f.Condition)
	}
	return f.Name + "=" + f.RawValue
}
