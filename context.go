package chi

import "context"

type contextKey struct{}

var RouteCtxKey = &contextKey{}

type Context struct {
	parentCtx context.Context
	URLParams RouteParams
}

type RouteParams struct {
	Keys   []string
	Values []string
}

func NewRouteContext() *Context {
	return &Context{}
}

func (x *Context) Reset() {
	x.parentCtx = nil
	x.URLParams.Keys = x.URLParams.Keys[:0]
	x.URLParams.Values = x.URLParams.Values[:0]
}

func (x *Context) WithParent(ctx context.Context) {
	x.parentCtx = ctx
}

func (x *Context) URLParam(key string) string {
	for i := len(x.URLParams.Keys) - 1; i >= 0; i-- {
		if x.URLParams.Keys[i] == key {
			return x.URLParams.Values[i]
		}
	}
	return ""
}

func (x *Context) pushURLParam(key string, value string) {
	x.URLParams.Keys = append(x.URLParams.Keys, key)
	x.URLParams.Values = append(x.URLParams.Values, value)
}

func RouteContext(ctx context.Context) *Context {
	if ctx == nil {
		return nil
	}
	if rctx, ok := ctx.Value(RouteCtxKey).(*Context); ok {
		return rctx
	}
	return nil
}

func contextWithRouteContext(parent context.Context, rctx *Context) context.Context {
	rctx.WithParent(parent)
	return context.WithValue(parent, RouteCtxKey, rctx)
}

func URLParam(r interface{ Context() context.Context }, key string) string {
	rctx := RouteContext(r.Context())
	if rctx == nil {
		return ""
	}
	return rctx.URLParam(key)
}
