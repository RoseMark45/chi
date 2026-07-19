package chi

import (
	"context"
	"sync"
)

var contextPool = &sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

type Context struct {
	URLParams   map[string]string
	middlewares []func(http.Handler) http.Handler
}

func NewContext() *Context {
	return contextPool.Get().(*Context)
}

func (c *Context) Reset() {
	for k := range c.URLParams {
		delete(c.URLParams, k)
	}
	c.middlewares = c.middlewares[:0]
}

func (c *Context) Release() {
	c.Reset()
	contextPool.Put(c)
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
