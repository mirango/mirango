package mirango

type Handler interface {
	ServeHTTP(*Context) interface{}
}

type HandlerFunc func(*Context) interface{}

func (f HandlerFunc) ServeHTTP(c *Context) interface{} {
	return f(c)
}

