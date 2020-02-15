package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/globusdigital/feature-toggles/api"
	"github.com/globusdigital/feature-toggles/messaging"
	"github.com/globusdigital/feature-toggles/storage"
)

type options struct {
	addr string

	storage storageKind
	mongodb string

	messaging messagingKind
	nats      string
}

type storageKind storage.Kind

func (v *storageKind) String() string {
	return storage.Kind(*v).String()
}

func (v *storageKind) Set(s string) error {
	switch s {
	case storage.MongoKind.String():
		*v = storageKind(storage.MongoKind)
	case storage.MemKind.String(), "":
		*v = storageKind(storage.MemKind)
	default:
		return errors.New("unknown storage type")
	}
	return nil
}

type messagingKind messaging.Kind

func (v *messagingKind) String() string {
	return messaging.Kind(*v).String()
}

func (v *messagingKind) Set(s string) error {
	switch s {
	case messaging.NatsKind.String():
		*v = messagingKind(messaging.NatsKind)
	case messaging.NoopKind.String(), "":
		*v = messagingKind(messaging.NoopKind)
	default:
		return errors.New("unknown messaging type")
	}
	return nil
}

func (o options) Store(ctx context.Context) (storage.Store, error) {
	switch storage.Kind(o.storage) {
	case storage.MemKind:
		return storage.NewMem(), nil
	case storage.MongoKind:
		return storage.NewMongo(ctx, o.mongodb)
	}

	return nil, errors.New("unknown storage type")
}

func (o options) Bus() (messaging.Bus, error) {
	switch messaging.Kind(o.messaging) {
	case messaging.NoopKind:
		return messaging.NewNoop(), nil
	case messaging.NatsKind:
		return messaging.NewNats(o.nats)
	}

	return nil, errors.New("unknown messaging type")
}

var (
	opts = options{}
)

func main() {
	flag.Parse()

	mainCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGQUIT)

	ctx, cancel := context.WithTimeout(mainCtx, time.Minute)
	defer cancel()
	store, err := opts.Store(ctx)
	if err != nil {
		log.Fatalf("Error initializing store")
	}

	bus, err := opts.Bus()
	if err != nil {
		log.Fatalf("Error initializing messaging bus")
	}

	server := &http.Server{Addr: opts.addr, Handler: api.Handler(store, bus)}
	go func() {
		log.Printf("Starting server on %s", opts.addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Error starting http server: %v", err)
		}
	}()

	<-c
	log.Println("Terminating server")
	ctx, cancel = context.WithTimeout(mainCtx, time.Minute)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("Error shutting down server:", err)
	}
	cancel()
}

func init() {
	flag.StringVar(&opts.addr, "addr", ":80", "listening address")
	flag.Var(&opts.storage, "storage", `storage type. Choices: mongo, mem (default "mem")`)
	flag.StringVar(&opts.mongodb, "mongodb", "mongodb://127.0.0.1:27017/featuretoggles", "mongodb address")
	flag.Var(&opts.messaging, "messaging", `messaging type. Choices: nats, noop (default "noop")`)
	flag.StringVar(&opts.nats, "nats", "nats://127.0.0.1:4222", "nats address")
}
