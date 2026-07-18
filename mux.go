package chi

import (
	"net/http"
	"strings"
	"sync"
)

type Router interface {
	Use(middlewares ...func(http.Handler) http.Handler)
	Get(pattern string, handlerFn http.HandlerFunc)
	Route(pattern string, fn func(r Router))
	http.Handler
}

type Mux struct {
	pool       sync.Pool
	middleware []func(http.Handler) http.Handler
	routes     []route
	prefix     string
}

type route struct {
	method  string
	pattern string
	handler http.Handler
}

func NewRouter() *Mux {
	mx := &Mux{}
	mx.pool.New = func() any {
		return NewRouteContext()
	}
	return mx
}

func (mx *Mux) Use(middlewares ...func(http.Handler) http.Handler) {
	mx.middleware = append(mx.middleware, middlewares...)
}

func (mx *Mux) Get(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(http.MethodGet, pattern, handlerFn)
}

func (mx *Mux) Route(pattern string, fn func(r Router)) {
	sub := &Mux{
		pool:       mx.pool,
		middleware: append([]func(http.Handler) http.Handler{}, mx.middleware...),
		prefix:     joinPatterns(mx.prefix, pattern),
	}
	fn(sub)
	mx.routes = append(mx.routes, sub.routes...)
}

func (mx *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rctx := RouteContext(r.Context())
	ownedRouteContext := false
	if rctx == nil {
		rctx = mx.pool.Get().(*Context)
		rctx.Reset()
		ownedRouteContext = true
		r = r.WithContext(contextWithRouteContext(r.Context(), rctx))
	}

	if ownedRouteContext {
		defer mx.pool.Put(rctx)
	}

	for _, route := range mx.routes {
		if route.method != r.Method {
			continue
		}
		if !matchPattern(route.pattern, r.URL.Path, rctx) {
			continue
		}
		route.handler.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

func (mx *Mux) handle(method string, pattern string, handler http.Handler) {
	fullPattern := joinPatterns(mx.prefix, pattern)
	for i := len(mx.middleware) - 1; i >= 0; i-- {
		handler = mx.middleware[i](handler)
	}
	mx.routes = append(mx.routes, route{
		method:  method,
		pattern: fullPattern,
		handler: handler,
	})
}

func joinPatterns(left string, right string) string {
	switch {
	case left == "":
		return right
	case right == "":
		return left
	case strings.HasSuffix(left, "/") && strings.HasPrefix(right, "/"):
		return left + strings.TrimPrefix(right, "/")
	case !strings.HasSuffix(left, "/") && !strings.HasPrefix(right, "/"):
		return left + "/" + right
	default:
		return left + right
	}
}

func matchPattern(pattern string, path string, rctx *Context) bool {
	patternParts := splitPath(pattern)
	pathParts := splitPath(path)
	if len(patternParts) != len(pathParts) {
		return false
	}

	startLenKeys := len(rctx.URLParams.Keys)
	startLenValues := len(rctx.URLParams.Values)

	for i, patternPart := range patternParts {
		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			key := strings.TrimSuffix(strings.TrimPrefix(patternPart, "{"), "}")
			rctx.pushURLParam(key, pathParts[i])
			continue
		}
		if patternPart != pathParts[i] {
			rctx.URLParams.Keys = rctx.URLParams.Keys[:startLenKeys]
			rctx.URLParams.Values = rctx.URLParams.Values[:startLenValues]
			return false
		}
	}

	return true
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}
