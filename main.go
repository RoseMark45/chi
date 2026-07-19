package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

var RouteCtxKey = &struct{}{}

type Context struct {
	URLParams map[string]string
}

func (c *Context) Reset() {
	for k := range c.URLParams {
		delete(c.URLParams, k)
	}
}

func RouteContext(ctx context.Context) *Context {
	if ctx == nil {
		return nil
	}
	if rc, ok := ctx.Value(RouteCtxKey).(*Context); ok {
		return rc
	}
	return nil
}

func URLParam(r *http.Request, key string) string {
	if r == nil {
		return ""
	}
	ctx := RouteContext(r.Context())
	if ctx == nil {
		return ""
	}
	return ctx.URLParams[key]
}

type Router struct {
	pool       sync.Pool
	middleware []func(http.Handler) http.Handler
	routes     []route
}

type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		pool: sync.Pool{
			New: func() interface{} {
				return &Context{URLParams: make(map[string]string)}
			},
		},
	}
}

func (r *Router) Use(m ...func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, m...)
}

func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.routes = append(r.routes, route{method: "GET", pattern: pattern, handler: handler})
}

func (r *Router) Route(prefix string, fn func(sub *Router)) {
	sub := NewRouter()
	fn(sub)
	for _, rt := range sub.routes {
		r.routes = append(r.routes, route{
			method:  rt.method,
			pattern: prefix + rt.pattern,
			handler: rt.handler,
		})
		// Merge sub-router middleware
		r.middleware = append(r.middleware, sub.middleware...)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rctx := r.pool.Get().(*Context)
	rctx.Reset()

	ctx := context.WithValue(req.Context(), RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	var handler http.HandlerFunc
	for _, rt := range r.routes {
		if rt.method != req.Method {
			continue
		}
		params, ok := matchRoute(rt.pattern, req.URL.Path)
		if ok {
			rctx.URLParams = params
			handler = rt.handler
			break
		}
	}

	if handler == nil {
		r.pool.Put(rctx)
		http.NotFound(w, req)
		return
	}

	final := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), RouteCtxKey, rctx)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))

	for i := len(r.middleware) - 1; i >= 0; i-- {
		final = r.middleware[i](final)
	}

	final.ServeHTTP(w, req)
	r.pool.Put(rctx)
}

func matchRoute(pattern, path string) (map[string]string, bool) {
	params := make(map[string]string)
	pi, si := 0, 0
	for pi < len(pattern) && si < len(path) {
		if pattern[pi] == '{' {
			pi++
			var name string
			for pi < len(pattern) && pattern[pi] != '}' {
				name += string(pattern[pi])
				pi++
			}
			pi++
			var value string
			for si < len(path) && path[si] != '/' {
				value += string(path[si])
				si++
			}
			params[name] = value
		} else if pattern[pi] == path[si] {
			pi++
			si++
		} else {
			return nil, false
		}
	}
	if pi != len(pattern) || si != len(path) {
		return nil, false
	}
	return params, true
}

func main() {
	fmt.Println("=== Chi URL Param Fix ===")
	fmt.Println("Fix: Ensure URL params survive middleware context replacement")
	fmt.Println()

	r := NewRouter()
	r.Route("/users/{userID}", func(sub *Router) {
		sub.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				ctx := context.WithValue(req.Context(), "customKey", "customValue")
				next.ServeHTTP(w, req.WithContext(ctx))
			})
		})
		sub.Get("/profile", func(w http.ResponseWriter, req *http.Request) {
			userID := URLParam(req, "userID")
			fmt.Fprintf(w, "userID=%s", userID)
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/users/123/profile")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("Response: %s\n", body)
	fmt.Printf("Status: %d\n", res.StatusCode)

	if string(body) == "userID=123" {
		fmt.Println("\n✓ TEST PASSED: URL params preserved through middleware context replacement")
	} else {
		fmt.Println("\n✗ TEST FAILED: URL params lost")
	}
}
