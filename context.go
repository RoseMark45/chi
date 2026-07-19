package main

import (
    "context"
)

// RouteCtxKey is a private key for the routing context.
type RouteCtxKey struct{}

// Context represents the routing context.
type Context struct {
    //... existing fields
    refCount int
}

// RouteContext returns the routing context from the given context.
func RouteContext(ctx context.Context) *Context {
    if ctx == nil {
        return &Context{}
    }
    if rc, ok := ctx.Value(RouteCtxKey{}).(*Context); ok {
        return rc.CloneContext()
    }
    return AcquireContext()
}

// WithRouteContext returns a new context with the provided routing context.
func WithRouteContext(parent context.Context, routeCtx *Context) context.Context {
    return context.WithValue(parent, RouteCtxKey{}, routeCtx.CloneContext())
}
