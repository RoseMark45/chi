package chi

import (
	"context"
	"net/http"
)

var RouteCtxKey = &contextKey{"RouteContext"}

type contextKey struct {
	name string
}

type Mux struct {
	pool     sync.Pool
	middles  []func(http.Handler) http.Handler
	routes   []route
}

type route struct {
	pattern string
	handler http.Handler
}

func NewRouter() *Mux {
	return &Mux{
		pool: sync.Pool{
			New: func() interface{} {
				return &Context{}
			},
		},
	}
}

func (mx *Mux) Use(middlewares ...func(http.Handler) http.Handler) {
	mx.middles = append(mx.middles, middlewares...)
}

func (mx *Mux) Get(pattern string, handler http.HandlerFunc) {
	mx.handle("GET", pattern, handler)
}

func (mx *Mux) handle(method, pattern string, handler http.Handler) {
	mx.routes = append(mx.routes, route{pattern: pattern, handler: handler})
}

func (mx *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rctx := mx.pool.Get().(*Context)
	rctx.Reset()
	defer mx.pool.Put(rctx)

	ctx := context.WithValue(r.Context(), RouteCtxKey, rctx)
	r = r.WithContext(ctx)

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, route := range mx.routes {
			params, ok := matchRoute(route.pattern, r.URL.Path)
			if ok {
				rctx.URLParams = params
				// Ensure we re-wrap the context so middleware replacements propagate
				ctx := context.WithValue(r.Context(), RouteCtxKey, rctx)
				route.handler.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		http.NotFound(w, r)
	})

	for i := len(mx.middles) - 1; i >= 0; i-- {
		handler = mx.middles[i](handler)
	}
	handler.ServeHTTP(w, r)
}

func matchRoute(pattern, path string) (map[string]string, bool) {
	params := make(map[string]string)
	// Simple route matching - handle {param} patterns
	pi, si := 0, 0
	for pi < len(pattern) && si < len(path) {
		if pattern[pi] == '{' {
			pi++
			var paramName string
			for pi < len(pattern) && pattern[pi] != '}' {
				paramName += string(pattern[pi])
				pi++
			}
			pi++
			var paramValue string
			for si < len(path) && path[si] != '/' {
				paramValue += string(path[si])
				si++
			}
			params[paramName] = paramValue
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

// Deprecated compatibility stubs
func (mx *Mux) Route(prefix string, fn func(r Router)) Router {
	return &RouterGroup{mx: mx, prefix: prefix}
}

type Router interface {
	Use(middlewares ...func(http.Handler) http.Handler)
	Get(pattern string, handler http.HandlerFunc)
}

type RouterGroup struct {
	mx     *Mux
	prefix string
}

func (g *RouterGroup) Use(middlewares ...func(http.Handler) http.Handler) {
	g.mx.Use(middlewares...)
}

func (g *RouterGroup) Get(pattern string, handler http.HandlerFunc) {
	panic("nested routing not supported in stub")
}
