package main

import (
    "sync"
)

var contextPool = sync.Pool{
    New: func() interface{} {
        return &Context{refCount: 0}
    },
}

// Context represents the routing context.
type Context struct {
    //... existing fields
    refCount int
}

// AcquireContext acquires a context from the pool.
func AcquireContext() *Context {
    ctx := contextPool.Get().(*Context)
    ctx.refCount = 1
    return ctx
}

// ReleaseContext releases the context back to the pool if it is not referenced.
func (c *Context) ReleaseContext() {
    c.refCount--
    if c.refCount == 0 {
        contextPool.Put(c)
    }
}

// CloneContext clones the context and increments the reference count.
func (c *Context) CloneContext() *Context {
    c.refCount++
    return c
}
