package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/globusdigital/feature-toggles/toggle"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type EventBus interface {
	Send(ctx context.Context, event toggle.Event) error
}

type Store interface {
	Get(ctx context.Context, serviceName string) ([]toggle.Flag, error)
	Save(ctx context.Context, flags []toggle.Flag, initial bool) error
	Delete(ctx context.Context, flags []toggle.Flag) error
}

func Handler(store Store, bus EventBus) http.Handler {
	r := chi.NewMux()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/flags", func(r chi.Router) {
		r.With(middleware.Timeout(time.Second*2)).Get("/", getAllFlags(store))

		r.Route("/{serviceName}", func(r chi.Router) {
			r.With(middleware.Timeout(time.Second*2)).Get("/", getFlags(store))
			r.With(middleware.Timeout(time.Second*10), flagsCtx).Post("/", saveFlags(store, bus))
			r.With(middleware.Timeout(time.Second*10), flagsCtx).Delete("/", deleteFlags(store, bus))

			r.Route("/initial", func(r chi.Router) {
				r.With(middleware.Timeout(time.Second*12), flagsCtx).Post("/", saveInitialFlags(store))
			})
		})
	})

	return r
}

type flagsCtxType string

var flagsKey flagsCtxType = "flags"

func flagsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var flags []toggle.Flag
		if err := json.Unmarshal(b, &flags); err != nil {
			status := http.StatusInternalServerError
			e := &json.SyntaxError{}
			if errors.As(err, &e) {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}

		if len(flags) == 0 {
			http.Error(w, "No flags given", http.StatusBadRequest)
			return
		}

		serviceName := chi.URLParam(r, "serviceName")

		for i, f := range flags {
			f = f.Normalized()
			flags[i] = f

			if f.ServiceName != serviceName && f.ServiceName != "" {
				http.Error(w, fmt.Sprintf("Invalid flag: %v", f), http.StatusBadRequest)
				return
			}

			if err := f.Condition.Validate(); err != nil {
				http.Error(w, fmt.Sprintf("Invalid flag %s condition: %v", f, err), http.StatusBadRequest)
				return
			}
		}

		ctx := context.WithValue(r.Context(), flagsKey, flags)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getAllFlags(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		getFlagsForServiceName(r.Context(), "", store, w)
	}
}

func getFlags(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceName := chi.URLParam(r, "serviceName")
		getFlagsForServiceName(r.Context(), serviceName, store, w)
	}
}

func saveFlags(store Store, bus EventBus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		flags := getFlagsFromCtx(ctx)
		if saveFlagsForService(ctx, flags, false, store, w, r.Body) {
			return
		}
		if err := bus.Send(ctx, toggle.Event{Type: toggle.SaveEvent, Flags: flags}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func saveInitialFlags(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		flags := getFlagsFromCtx(ctx)
		if saveFlagsForService(ctx, flags, true, store, w, r.Body) {
			return
		}
		serviceName := chi.URLParam(r, "serviceName")
		getFlagsForServiceName(r.Context(), serviceName, store, w)
	}
}

func deleteFlags(store Store, bus EventBus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		flags := getFlagsFromCtx(ctx)
		if saveFlagsForService(ctx, flags, false, store, w, r.Body) {
			return
		}
		if err := bus.Send(ctx, toggle.Event{Type: toggle.DeleteEvent, Flags: flags}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getFlagsForServiceName(ctx context.Context, serviceName string, store Store, w http.ResponseWriter) {
	flags, err := store.Get(ctx, serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(flags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(b)
}

func saveFlagsForService(ctx context.Context, flags []toggle.Flag, initial bool, store Store, w http.ResponseWriter, r io.Reader) bool {
	if err := store.Save(ctx, flags, initial); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}

	return false
}

func getFlagsFromCtx(ctx context.Context) []toggle.Flag {
	if flags, ok := ctx.Value(flagsKey).([]toggle.Flag); ok {
		return flags
	}

	return nil
}
