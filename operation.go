package mirango

import (
	"fmt"

	"github.com/mirango/defaults"
	"github.com/mirango/framework"
)

type Operations struct {
	operations []*Operation
	index      map[string]int
}

func NewOperations() *Operations {
	return &Operations{
		index: map[string]int{},
	}
}

func (ops *Operations) Append(operations ...*Operation) {
	le := len(ops.operations)
	for i := 0; i < len(operations); i++ {
		name := operations[i].name
		if name == "" {
			ops.operations = append(ops.operations, operations[i])
			continue
		} else {
			if _, ok := ops.index[name]; ok {
				panic(fmt.Sprintf("Detected 2 operations with the same name: \"%s\".", name))
			}
			ops.operations = append(ops.operations, operations[i])
			ops.index[name] = le + i
		}
	}
}

func (ops *Operations) Set(operations ...*Operation) {
	ops.operations = nil
	ops.index = map[string]int{}
	ops.Append(operations...)
}

func (ops *Operations) Get(name string) *Operation {
	return ops.operations[ops.index[name]]
}

func (ops *Operations) GetAll() []*Operation {
	return ops.operations
}

func (ops *Operations) Union(nops *Operations) {
	for _, o := range nops.operations {
		ops.Append(o)
	}
}

func (ops *Operations) Clone() *Operations {
	nops := NewOperations()
	nops.Union(ops)
	return nops
}

func (ops *Operations) GetByIndex(i int) *Operation {
	return ops.operations[i]
}

func (ops *Operations) GetByMethod(method string) *Operation {
	for _, o := range ops.operations {
		for _, m := range o.methods {
			if m == method {
				return o
			}
		}
	}
	return nil
}

func (ops *Operations) Len() int {
	return len(ops.operations)
}

type Operation struct {
	name          string
	handler       Handler
	route         *Route
	methods       []string
	schemes       []string
	accepts       []string
	returns       []string
	middleware    []Middleware
	params        *Params
	renders       string
	mimeTypeIn    paramIn
	mimeTypeParam string

	returnsOnly bool
	acceptsOnly bool
	schemesOnly bool
}

func NewOperation(h interface{}) *Operation {
	handler, err := handler(h)
	if err != nil {
		panic(err.Error())
	}
	o := &Operation{
		methods: []string{"GET"},
		handler: handler,
		params:  NewParams(),
		schemes: defaults.Schemes,
		accepts: defaults.Accepts,
		returns: defaults.Returns,
	}
	o.middleware = []Middleware{CheckReturns(o), CheckSchemes(o), CheckAccepts(o), CheckParams(o)}
	return o
}

func GET(h interface{}) *Operation {
	return NewOperation(h).Methods("GET")
}

func POST(h interface{}) *Operation {
	return NewOperation(h).Methods("POST")
}

func PUT(h interface{}) *Operation {
	return NewOperation(h).Methods("PUT")
}

func DELETE(h interface{}) *Operation {
	return NewOperation(h).Methods("DELETE")
}

func (o *Operation) Uses(h interface{}) *Operation { //interface
	handler, err := handler(h)
	if err != nil {
		panic(err.Error())
	}
	o.handler = handler
	return o
}

func getHandler(h interface{}, mw []Middleware) (Handler, error) {
	final, err := handler(h)
	if err != nil {
		return nil, err
	}
	for i := len(mw) - 1; i >= 0; i-- {
		final = mw[i].Run(final)
	}
	return final, nil
}

func (o *Operation) With(mw ...interface{}) *Operation {
	for i := 0; i < len(mw); i++ {
		switch t := mw[i].(type) {
		case Middleware:
			o.middleware = middlewareAppend(o.middleware, t)
		case MiddlewareFunc:
			o.middleware = middlewareAppend(o.middleware, t)
		case func(Handler) Handler:
			o.middleware = middlewareAppend(o.middleware, MiddlewareFunc(t))
		}
	}
	return o
}

func (o *Operation) with() {
	handler, err := getHandler(o.handler, o.middleware)
	if err != nil {
		panic(err.Error())
	}
	o.handler = handler
}

func (o *Operation) Apply(temps ...func(*Operation)) *Operation {
	for i := 0; i < len(temps); i++ {
		temps[i](o)
	}
	return o
}

func (o *Operation) GetRoute() framework.Route {
	return o.route
}

func (o *Operation) Name(name string) *Operation {
	o.name = name
	return o
}

func (o *Operation) GetName() string {
	return o.name
}

func (o *Operation) Methods(methods ...string) *Operation {
	o.methods = methods
	return o
}

func (o *Operation) Params(params ...*Param) *Operation {
	o.params.Set(params...) // make it append instead of set
	return o
}

func (o *Operation) GetParams() *Params {
	return o.params
}

func (o *Operation) getAllParams() {
	params := o.params.Clone()
	params.Union(o.route.params)
	o.params = params
}

func (o *Operation) Schemes(schemes ...string) *Operation {
	o.schemes = append(o.schemes, schemes...)
	o.schemesOnly = false
	return o
}

func (o *Operation) Accepts(accepts ...string) *Operation {
	o.accepts = append(o.accepts, accepts...)
	o.acceptsOnly = false
	return o
}

func (o *Operation) Returns(returns ...string) *Operation {
	o.returns = append(o.returns, returns...)
	o.returnsOnly = false
	return o
}

func (o *Operation) SchemesOnly(schemes ...string) *Operation {
	o.schemes = schemes
	o.schemesOnly = true
	return o
}

func (o *Operation) AcceptsOnly(accepts ...string) *Operation {
	o.accepts = accepts
	o.acceptsOnly = true
	return o
}

func (o *Operation) ReturnsOnly(returns ...string) *Operation {
	o.returns = returns
	o.returnsOnly = true
	return o
}

func (o *Operation) PathParam(name string) *Param {
	p := PathParam(name)
	o.params.Append(p)
	return p
}

func (o *Operation) QueryParam(name string) *Param {
	p := QueryParam(name)
	o.params.Append(p)
	return p
}

func (o *Operation) HeaderParam(name string) *Param {
	p := HeaderParam(name)
	o.params.Append(p)
	return p
}

func (o *Operation) BodyParam(name string) *Param {
	p := BodyParam(name)
	o.params.Append(p)
	return p
}

func (o *Operation) GetMethods() []string {
	return o.methods
}

func (o *Operation) GetMiddleware() []Middleware {
	return o.middleware
}

func (o *Operation) getAllMiddleware() {
	o.middleware = middlewareUnion(o.middleware, o.route.middleware)
}

func (o *Operation) GetSchemes() []string {
	return o.methods
}

func (o *Operation) getAllSchemes() {
	schemes := o.schemes
	if !o.schemesOnly {
		schemes = stringsUnion(schemes, o.route.schemes)
	}
	o.schemes = schemes
}

func (o *Operation) GetAccepts() []string {
	return o.accepts
}

func (o *Operation) getAllAccepts() {
	accepts := o.accepts
	if !o.acceptsOnly {
		accepts = stringsUnion(accepts, o.route.accepts)
	}
	o.accepts = accepts
}

func (o *Operation) GetReturns() []string {
	return o.returns
}

func (o *Operation) getAllReturns() {
	returns := o.returns
	if !o.returnsOnly {
		returns = stringsUnion(returns, o.route.returns)
	}
	o.returns = returns
}

func (o *Operation) BuildPath(v ...interface{}) string {
	return o.route.BuildPath(v...)
}

func (o *Operation) GetPath() string {
	return o.route.path
}

func (o *Operation) GetFullPath() string {
	return o.route.GetFullPath()
}

func (o *Operation) ServeHTTP(c *Context) interface{} {
	c.operation = o
	return o.handler.ServeHTTP(c)
}

func (o *Operation) Clone() *Operation {
	no := NewOperation(o.handler)

	no.methods = o.methods
	no.schemes = o.schemes
	no.accepts = o.accepts
	no.returns = o.returns
	no.middleware = o.middleware
	no.params = o.params.Clone()
	no.renders = o.renders
	no.mimeTypeIn = o.mimeTypeIn
	no.mimeTypeParam = o.mimeTypeParam

	return no
}

type middlewareContainer struct {
	middleware []interface{}
}

func With(mw ...interface{}) *middlewareContainer {
	return &middlewareContainer{middleware: mw}
}

func (c middlewareContainer) Handle(operations ...*Operation) []*Operation {
	for i := 0; i < len(operations); i++ {
		operations[i].With(c.middleware...)
	}
	return operations
}

type templateContainer struct {
	templates []func(*Operation)
}

func Apply(temps ...func(*Operation)) *templateContainer {
	return &templateContainer{templates: temps}
}

func (c templateContainer) To(operations ...*Operation) []*Operation {
	for i := 0; i < len(operations); i++ {
		operations[i].Apply(c.templates...)
	}
	return operations
}
