package main

import (
	"context"
	"net/http"
)

// RouteCtxKey is the private key under which *Context is stored in the
// request context. Exported as a string so downstream code can recover it
// even when a middleware wraps the request context with WithContext.
type ctxKey struct{}

var RouteCtxKey = ctxKey{}

// Context holds the routing state (URL params, route pattern) for a request.
type Context struct {
	Params    map[string]string
	RoutePath string
}

// RouteContext recovers the active routing *Context from any request context
// chain. If the context was wrapped (e.g. ctx = context.WithValue(req.Context(), ...)
// or req.WithContext(newCtx)), we still walk up the chain to find the *Context.
func RouteContext(ctx context.Context) *Context {
	if ctx == nil {
		return nil
	}
	if rc, ok := ctx.Value(RouteCtxKey).(*Context); ok {
		return rc
	}
	return nil
}

// URLParam returns the named URL parameter from the active routing context.
func URLParam(r *http.Request, key string) string {
	rc := RouteContext(r.Context())
	if rc == nil {
		return ""
	}
	return rc.Params[key]
}
