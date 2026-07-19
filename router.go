package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

// NewRouter creates a new router with the necessary middlewares.
func NewRouter() *chi.Mux {
    r := chi.NewRouter()

    // Apply the PreserveRouteContextMiddleware globally
    r.Use(PreserveRouteContextMiddleware)

    // Add your routes here
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Welcome!"))
    })

    r.Route("/sub", func(r chi.Router) {
        r.Get("/", func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte("Sub route!"))
        })
    })

    return r
}
