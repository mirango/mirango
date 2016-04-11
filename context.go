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
	values    framework.Values
	ended     bool
}

func NewContext(res *Response, req *Request) *Context {
	return &Context{
		Response: res,
		Request:  req,
		values:   framework.Values{},
	}
}

func (c *Context) Operation() framework.Operation {
	return c.operation
}

func (c *Context) Values() framework.Values {
	return c.values
}

func (c *Context) SetValue(k interface{}, v interface{}) {
	c.values.Set(k, v)
}

func (c *Context) GetValue(k interface{}) framework.Value {
	return c.values.Get(k)
}

func (c *Context) Value(k interface{}) interface{} {
	return c.values[k]
}

func (c *Context) Header() http.Header {
	return c.Response.Header()
}

func (c *Context) End() {
	c.ended = true
}
