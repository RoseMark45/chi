package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

// PreserveRouteContextMiddleware is a middleware that preserves the routing context when the request is cloned.
func PreserveRouteContextMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get the current routing context
        routeCtx := chi.RouteContext(r.Context())

        // Create a new context with the preserved routing context
        newCtx := chi.WithRouteContext(r.Context(), routeCtx)

        // Clone the request with the new context
        newR := r.WithContext(newCtx)

        // Call the next handler with the cloned request
        next.ServeHTTP(w, newR)
    })
}
