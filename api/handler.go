package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/globusdigital/feature-toggles/messaging"
	"github.com/globusdigital/feature-toggles/storage"
	"github.com/globusdigital/feature-toggles/toggle"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Handler(store storage.Store, bus messaging.Bus) http.Handler {
	r := chi.NewMux()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/flags", func(r chi.Router) {
		r.With(middleware.Timeout(time.Second*2)).Get("/", getAllFlags(store))

		r.Route("/{serviceName}", func(r chi.Router) {
			r.With(middleware.Timeout(time.Second*2)).Get("/", getFlags(store))
			r.With(middleware.Timeout(time.Second*10)).Post("/", saveFlags(store, bus))

			r.Route("/initial", func(r chi.Router) {
				r.With(middleware.Timeout(time.Second*12)).Post("/", saveInitialFlags(store))
			})
		})
	})

	return r
}

func getAllFlags(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		getFlagsForServiceName(r.Context(), "", store, w)
	}
}

func getFlags(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceName := chi.URLParam(r, "serviceName")
		getFlagsForServiceName(r.Context(), serviceName, store, w)
	}
}

func saveFlags(store storage.Store, bus messaging.Bus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceName := chi.URLParam(r, "serviceName")
		flags := saveFlagsForService(r.Context(), serviceName, false, store, w, r.Body)
		if err := bus.Send(r.Context(), flags); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func saveInitialFlags(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceName := chi.URLParam(r, "serviceName")
		saveFlagsForService(r.Context(), serviceName, true, store, w, r.Body)
		getFlagsForServiceName(r.Context(), serviceName, store, w)
	}
}

func getFlagsForServiceName(ctx context.Context, serviceName string, store storage.Store, w http.ResponseWriter) {
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

	w.Write(b)
}

func saveFlagsForService(ctx context.Context, serviceName string, initial bool, store storage.Store, w http.ResponseWriter, r io.Reader) []toggle.Flag {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	var flags []toggle.Flag
	if err := json.Unmarshal(b, &flags); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	for _, f := range flags {
		if f.ServiceName != serviceName && f.ServiceName != "" {
			http.Error(w, fmt.Sprintf("Invalid flag: %v", f), http.StatusBadRequest)
			return nil
		}
	}

	if err := store.Save(ctx, flags, initial); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	return flags
}
