package mirango

import (
	"net/http"

	"github.com/mirango/framework"
)

type Context struct {
	framework.LogWriter
	*Response
	*Request
	operation *Operation
	id        int64
	values    framework.Values
	user      framework.User
	locale    framework.Locale
	ended     bool
}

func NewContext(res *Response, req *Request) *Context {
	return &Context{
		Response: res,
		Request:  req,
		values:   framework.Values{},
	}
}

func (c *Context) ID() int64 {
	return c.id
}

func (c *Context) Locale() framework.Locale {
	return c.locale
}

func (c *Context) User() framework.User {
	return c.user
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
