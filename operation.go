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
		for _, method := range operations[i].methods {
			if ops.GetByMethod(method) != nil {
				panic(fmt.Sprintf("Detected 2 operations with the same method: \"%s\".", method))
			}
		}
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
	for _, o := range ops.GetAll() {
		nops.Append(o.Clone())
	}
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

func (ops *Operations) Apply(p ...Preset) {
	for _, o := range ops.operations {
		o.Apply(p...)
	}
}

func (ops *Operations) finalize() {
	for _, o := range ops.operations {
		o.finalize()
	}
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
	return o
}

func GET(h interface{}) *Operation {
	return GETNested(h, nil)
}

func GETNested(h interface{}, cb func(*Operation)) *Operation {
	o := NewOperation(h).Methods("GET")

	if cb != nil {
		cb(o)
	}

	return o
}

func POST(h interface{}) *Operation {
	return POSTNested(h, nil)
}

func POSTNested(h interface{}, cb func(*Operation)) *Operation {
	o := NewOperation(h).Methods("POST")

	if cb != nil {
		cb(o)
	}

	return o
}

func PUT(h interface{}) *Operation {
	return PUTNested(h, nil)
}

func PUTNested(h interface{}, cb func(*Operation)) *Operation {
	o := NewOperation(h).Methods("PUT")

	if cb != nil {
		cb(o)
	}

	return o
}

func PATCH(h interface{}) *Operation {
	return PATCHNested(h, nil)
}

func PATCHNested(h interface{}, cb func(*Operation)) *Operation {
	o := NewOperation(h).Methods("PATCH")

	if cb != nil {
		cb(o)
	}

	return o
}

func DELETE(h interface{}) *Operation {
	return DELETENested(h, nil)
}

func DELETENested(h interface{}, cb func(*Operation)) *Operation {
	o := NewOperation(h).Methods("DELETE")

	if cb != nil {
		cb(o)
	}

	return o
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

func (o *Operation) Apply(p ...Preset) *Operation {
	for i := 0; i < len(p); i++ {
		p[i].ApplyTo(o)
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
	o.middleware = append(o.middleware, CheckReturns(o), CheckSchemes(o), CheckAccepts(o), CheckParams(o))
	o.middleware = middlewareUnion(o.middleware, o.route.middleware)
}

func (o *Operation) GetSchemes() []string {
	return o.methods
}

func (o *Operation) getAllSchemes() {
	if !o.schemesOnly {
		o.schemes = stringsUnion(o.schemes, o.route.schemes)
	}
}

func (o *Operation) GetAccepts() []string {
	return o.accepts
}

func (o *Operation) getAllAccepts() {
	if !o.acceptsOnly {
		o.accepts = stringsUnion(o.accepts, o.route.accepts)
	}
}

func (o *Operation) GetReturns() []string {
	return o.returns
}

func (o *Operation) getAllReturns() {
	if !o.returnsOnly {
		o.returns = stringsUnion(o.returns, o.route.returns)
	}
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

func (o *Operation) finalize() {
	o.getAllSchemes()
	o.getAllAccepts()
	o.getAllReturns()
	o.getAllParams()
	o.getAllMiddleware()
	o.Apply(o.route.presets...)
	o.with()
}

func (o *Operation) Clone() *Operation {
	no := NewOperation(o.handler)

	no.methods = o.methods
	no.schemes = o.schemes
	no.accepts = o.accepts
	no.returns = o.returns
	no.middleware = o.middleware
	no.params = o.params.Clone()
	no.mimeTypeIn = o.mimeTypeIn
	no.mimeTypeParam = o.mimeTypeParam
	no.handler = o.handler
	no.returnsOnly = o.returnsOnly
	no.acceptsOnly = o.acceptsOnly
	no.schemesOnly = o.schemesOnly

	return no
}

type middleware []interface{}

func With(mw ...interface{}) middleware {
	return middleware(mw)
}

func (mw middleware) Handle(operations ...*Operation) {
	for i := 0; i < len(operations); i++ {
		operations[i].With(mw...)
	}
}

type Preset interface {
	ApplyTo(*Operation)
}

type PresetFunc func(*Operation)

func (f PresetFunc) ApplyTo(o *Operation) {
	f(o)
}

type Presets []Preset

func Apply(p ...Preset) Presets {
	return Presets(p)
}

func (p Presets) To(operations ...*Operation) {
	for i := 0; i < len(operations); i++ {
		operations[i].Apply(p...)
	}
}

func (p Presets) Union(o Presets) {
	for _, oo := range o {
		exists := false
		for _, pp := range p {
			if pp == oo {
				exists = true
				break
			}
		}
		if !exists {
			p = append(p, oo)
		}
	}
}
