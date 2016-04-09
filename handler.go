package mirango

import (
	"fmt"
	"net/http"

	"github.com/wlMalk/mirango/framework"
)

type Handler interface {
	ServeHTTP(*Context) interface{}
}

type HandlerFunc func(*Context) interface{}

func (f HandlerFunc) ServeHTTP(c *Context) interface{} {
	return f(c)
}

func handler(h interface{}) (Handler, error) {
	if h == nil {
		return nil, fmt.Errorf("handler is nil")
	}
	switch t := h.(type) {
	case Handler:
		return t, nil
	case HandlerFunc:
		return t, nil
	case func(*Context) interface{}:
		return HandlerFunc(t), nil
	case func(*Context):
		return HandlerFunc(func(c *Context) interface{} {
			t(c)
			c.End()
			return nil
		}), nil
	case framework.Handler:
		return HandlerFunc(func(c *Context) interface{} {
			return t.ServeHTTP(c)
		}), nil
	case framework.HandlerFunc:
		return HandlerFunc(func(c *Context) interface{} {
			return t.ServeHTTP(c)
		}), nil
	case func(framework.Context) interface{}:
		return HandlerFunc(func(c *Context) interface{} {
			return t(c)
		}), nil
	case func(framework.Context):
		return HandlerFunc(func(c *Context) interface{} {
			t(c)
			c.End()
			return nil
		}), nil
	case func(*Response, *Request) interface{}:
		return HandlerFunc(func(c *Context) interface{} {
			return t(c.Response, c.Request)
		}), nil
	case func(*Response, *Request):
		return HandlerFunc(func(c *Context) interface{} {
			t(c.Response, c.Request)
			c.End()
			return nil
		}), nil
	case func(framework.Response, framework.Request) interface{}:
		return HandlerFunc(func(c *Context) interface{} {
			return t(c.Response, c.Request)
		}), nil
	case func(framework.Response, framework.Request):
		return HandlerFunc(func(c *Context) interface{} {
			t(c.Response, c.Request)
			c.End()
			return nil
		}), nil
	case http.Handler:
		return HandlerFunc(func(c *Context) interface{} {
			t.ServeHTTP(c.Response, c.Request.Request)
			c.End()
			return nil
		}), nil
	case http.HandlerFunc:
		return HandlerFunc(func(c *Context) interface{} {
			t.ServeHTTP(c.Response, c.Request.Request)
			c.End()
			return nil
		}), nil
	case func(http.ResponseWriter, *http.Request) interface{}:
		return HandlerFunc(func(c *Context) interface{} {
			return t(c.Response, c.Request.Request)
		}), nil
	case func(http.ResponseWriter, *http.Request):
		return HandlerFunc(func(c *Context) interface{} {
			t(c.Response, c.Request.Request)
			c.End()
			return nil
		}), nil
	default:
		return nil, fmt.Errorf("wrong type of handler (%T)", h)
	}
}
