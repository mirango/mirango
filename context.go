package mirango

import (
	"net/http"

	"github.com/wlMalk/mirango/framework"
)

type Context struct {
	framework.LogWriter
	*Response
	*Request
	operation *Operation
	ended     bool
}

func NewContext(res *Response, req *Request) *Context {
	return &Context{
		Response: res,
		Request:  req,
	}
}

func (c *Context) Operation() framework.Operation {
	return c.operation
}

func (c *Context) Header() http.Header {
	return c.Response.Header()
}

func (c *Context) End() {
	c.ended = true
}
