package toggle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
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
	o := (clientOptions{
		values:         []ConditionValue{{Name: ServiceNameValue, Type: StringType, Value: name}},
		updateDuration: 30 * time.Minute,
		httpClient:     http.DefaultClient,
		log:            log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile),
		path:           "/flags",
	}).Apply(opts)

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
		if f.ServiceName != c.name && (f.ServiceName != "" || !o.global) {
			continue
		}

		values := make([]ConditionValue, len(c.opts.values)+len(o.values))
		copy(values, c.opts.values)
		copy(values[len(c.opts.values):], o.values)
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
		defer close(errC)
		if err := c.seedFlags(ctx, addr); err != nil {
			errC <- err
			return
		}

		if ctx.Err() != nil {
			return
		}

		var ch <-chan Event
		if c.opts.eventBus != nil {
			var err error
			if ch, err = c.opts.eventBus.Receive(ctx); err != nil {
				c.opts.log.Println("Error initializing event bus receiver:", err)
			}
		}

		ticker := time.NewTicker(c.opts.updateDuration)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case ev, open := <-ch:
				if !open {
					ch = nil
				}
				c.processEvent(ev)
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

	r, err := http.NewRequestWithContext(ctx, "POST", addr+path.Join(c.opts.path, c.name, "initial"), bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("creating initial flag request: %v", err)
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := c.opts.httpClient.Do(r)
	if err != nil {
		return fmt.Errorf("getting initial flag response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code for %s: %d (%s)", cleanupURL(r.URL), resp.StatusCode, resp.Status)
	}

	return c.updateStore(resp.Body)
}

func (c *Client) pollFlags(ctx context.Context, addr string) error {
	c.opts.log.Println("Polling for flags")

	r, err := http.NewRequestWithContext(ctx, "GET", addr+path.Join(c.opts.path, c.name), nil)
	if err != nil {
		return fmt.Errorf("creating update poll flag request: %v", err)
	}

	resp, err := c.opts.httpClient.Do(r)
	if err != nil {
		return fmt.Errorf("getting update poll flag response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code for %s: %d (%s)", cleanupURL(r.URL), resp.StatusCode, resp.Status)
	}

	return c.updateStore(resp.Body)
}

func (c *Client) processEvent(ev Event) {
	c.opts.log.Printf("Processing event %q with flags: %s", ev.Type, ev.Flags)

	switch ev.Type {
	case SaveEvent, DeleteEvent:
		for i := range ev.Flags {
			ev.Flags[i] = ev.Flags[i].Normalized()
		}
	default:
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	switch ev.Type {
	case SaveEvent:
		for _, f := range ev.Flags {
			flags := c.store[f.Name]

			var found bool
			for i, stored := range flags {
				if stored.ServiceName == f.ServiceName {
					flags[i] = f
					found = true
					break
				}
			}

			if !found {
				c.store[f.Name] = append(c.store[f.Name], f)
			}
		}
	case DeleteEvent:
		for _, f := range ev.Flags {
			flags := c.store[f.Name]
			for i, stored := range flags {
				if stored.ServiceName == f.ServiceName {
					if len(flags) == 1 {
						delete(c.store, f.Name)
					} else {
						flags = append(flags[:i], flags[i+1:]...)
					}
					break
				}
			}
		}
	}
}

func (c *Client) updateStore(r io.Reader) error {
	var flags []Flag
	if err := json.NewDecoder(r).Decode(&flags); err != nil {
		return fmt.Errorf("decoding flag data: %v", err)
	}

	store := map[string][]Flag{}
	for _, f := range flags {
		f = f.Normalized()
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
	name := f.Name
	if f.ServiceName != "" {
		name += "[" + f.ServiceName + "]"
	}
	if f.Condition.hasMatchers() {
		return fmt.Sprintf("%s=%s %s", name, f.RawValue, f.Condition)
	}
	return name + "=" + f.RawValue
}

func (f Flag) Normalized() Flag {
	f.Name = normalizeName(f.Name)
	f.ServiceName = normalizeSerivceName(f.ServiceName)

	return f
}

func cleanupURL(u *url.URL) string {
	u.User = url.User(u.User.Username())
	return u.String()
}
