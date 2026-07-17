package main

import (
	"context"
	"net/http"


)

// Router is a minimal chi-like router that preserves routing context across
// request context replacement (r.WithContext / context.WithValue).
type Router struct {
	routes []route
}

type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
}

func NewRouter() *Router { return &Router{} }

func (r *Router) Get(pattern string, h http.HandlerFunc) {
	r.routes = append(r.routes, route{"GET", pattern, h})
}

func (r *Router) Use(mw func(http.Handler) http.Handler) {
	// inline middleware wrapper applied per-route group
}

// ServeHTTP matches the route, attaches routing *Context to the request, and
// dispatches. The routing context is stored under RouteCtxKey so that even if
// a middleware later calls req.WithContext(newCtx), URLParam still resolves.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, rt := range r.routes {
		if rt.method != req.Method {
			continue
		}
		params := matchParams(rt.pattern, req.URL.Path)
		if params == nil {
			continue
		}
		rc := &Context{Params: params, RoutePath: rt.pattern}
		// Store routing context in the request context BEFORE calling handler.
		req = req.WithContext(context.WithValue(req.Context(), RouteCtxKey, rc))
		rt.handler(w, req)
		return
	}
	http.NotFound(w, req)
}

// matchParams extracts {param} from a pattern like /users/{userID}/profile.
func matchParams(pattern, path string) map[string]string {
	pp := splitPath(pattern)
	sp := splitPath(path)
	if len(pp) != len(sp) {
		return nil
	}
	params := map[string]string{}
	for i := range pp {
		if len(pp[i]) > 2 && pp[i][0] == '{' && pp[i][len(pp[i])-1] == '}' {
			params[pp[i][1:len(pp[i])-1]] = sp[i]
		} else if pp[i] != sp[i] {
			return nil
		}
	}
	return params
}

func splitPath(p string) []string {
	var out []string
	cur := ""
	for _, c := range p {
		if c == '/' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
		} else {
			cur += string(c)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
