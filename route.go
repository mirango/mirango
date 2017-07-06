package mirango

import (
	"net/http"
	"strings"

	"github.com/mirango/errors"
	"github.com/mirango/validation"
)

type Route struct {
	mirango *Mirango

	node *node

	parent *Route

	operations              *Operations
	path                    string
	schemes                 []string
	accepts                 []string
	returns                 []string
	middleware              []Middleware
	params                  *Params
	notFoundHandler         Handler
	methodNotAllowedHandler Handler
	panicHandler            Handler

	presets Presets

	returnsOnly bool
	acceptsOnly bool
	schemesOnly bool
}

func createNodes(parts []string, names []string, indices []int, typs []int, cs bool) *node {
	var pNode *node
	var node *node
	for i, part := range parts {
		node = newNode(part)
		node.caseSensitive = cs
		node.setParam(names[i], indices[i], typs[i] == 1)
		if pNode != nil {
			pNode.addNode(node)
		}
		pNode = node
	}
	return node
}

func NewRoute(path interface{}) *Route {
	return NewRouteNested(path, nil)
}

func NewRouteNested(path interface{}, cb func(*Route)) *Route {
	cs := false
	var nPath string

	switch p := path.(type) {
	case CaseSensitive:
		cs = true
		nPath = string(p)
	case string:
		nPath = p
	default:
		panic("invalid path")
	}

	slices := splitPath(nPath)
	parts, names, indices, typs := processPath(slices)
	node := createNodes(parts, names, indices, typs, cs)

	route := &Route{
		node:       node,
		path:       "/" + strings.Join(slices, "/"),
		operations: NewOperations(),
		params:     NewParams(),
	}

	node.route = route

	if cb != nil {
		cb(route)
	}

	return route
}

type CaseSensitive string

func (r *Route) Branch(path interface{}) *Route {
	return r.BranchNested(path, nil)
}

func (r *Route) BranchNested(path interface{}, cb func(*Route)) *Route {
	route := NewRoute(path)
	route = r.AddRoute(route)

	if cb != nil {
		cb(route)
	}

	return route
}

func (r *Route) AddRoute(route *Route) *Route {
	return r.AddRouteNested(route, nil)
}

func (r *Route) AddRouteNested(route *Route, cb func(*Route)) *Route {

	if route == nil {
		panic("route is nil")
	}

	if r.node.hasWildcard {
		panic("wildcard routes can not have sub-routes")
	}

	route = route.Clone()

	r.node.addNode(route.getTopNode())
	//r.node.compareNodes()

	route.parent = r

	route.mirango = r.mirango

	if cb != nil {
		cb(route)
	}

	return route
}

func (r *Route) getTopNode() *node {
	if r.parent == nil {
		return r.node.getRoot()
	}
	return r.parent.node.subtract(r.node)
}

func (r *Route) Clone() *Route {
	route := NewRoute(r.path)

	route.operations = r.operations.Clone()
	route.params = r.params.Clone()
	route.middleware = r.middleware
	route.schemes = r.schemes
	route.accepts = r.accepts
	route.returns = r.returns
	route.notFoundHandler = r.notFoundHandler
	route.methodNotAllowedHandler = r.methodNotAllowedHandler
	route.panicHandler = r.panicHandler
	route.presets = r.presets
	route.returnsOnly = r.returnsOnly
	route.acceptsOnly = r.acceptsOnly
	route.schemesOnly = r.schemesOnly

	for _, o := range route.operations.GetAll() {
		o.route = route
	}

	return route
}

func processPath(slices []string) (parts []string, names []string, indices []int, typs []int) {

	if len(slices) == 0 {
		panic("path is empty")
	}

	for i, s := range slices {
		index := -1
		typ := -1
		colon := strings.LastIndex(s, ":")
		wildcard := strings.LastIndex(s, "*")
		if colon > wildcard {
			index = colon
			typ = 0
		} else if colon < wildcard && i == len(slices)-1 {
			index = wildcard
			typ = 1
		}

		if index != -1 && index != len(s)-1 {
			parts = append(parts, s[:index])
			names = append(names, s[index+1:])
		} else {
			parts = append(parts, s)
			names = append(names, "")
		}
		indices = append(indices, index)
		typs = append(typs, typ)
	}

	return
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")

	if path == "" {
		return []string{""}
	}

	slices := strings.Split(path, "/")
	var nSlices []string

	for _, s := range slices {
		if len(s) > 0 {
			nSlices = append(nSlices, s)
		}
	}

	return nSlices
}

func (r *Route) GetNotFoundHandler() Handler {
	if r.parent != nil && r.notFoundHandler == nil {
		return r.parent.GetNotFoundHandler()
	}
	return r.notFoundHandler
}

func (r *Route) GetMethodNotAllowedHandler() Handler {
	if r.parent != nil && r.methodNotAllowedHandler == nil {
		return r.parent.GetMethodNotAllowedHandler()
	}
	return r.methodNotAllowedHandler
}

func (r *Route) GetPanicHandler() Handler {
	if r.parent != nil && r.panicHandler == nil {
		return r.parent.GetPanicHandler()
	}
	return r.panicHandler
}

func (r *Route) NotFoundHandler(h interface{}) *Route {
	handler, err := handler(h)
	if err != nil {
		panic(err.Error())
	}
	r.notFoundHandler = handler
	return r
}

func (r *Route) MethodNotAllowedHandler(h interface{}) *Route {
	handler, err := handler(h)
	if err != nil {
		panic(err.Error())
	}
	r.methodNotAllowedHandler = handler
	return r
}

func (r *Route) PanicHandler(h interface{}) *Route {
	handler, err := handler(h)
	if err != nil {
		panic(err.Error())
	}
	r.panicHandler = handler
	return r
}

func (r *Route) BuildPath(v ...interface{}) string {
	return ""
}

func (r *Route) GetPath() string {
	return r.path
}

func (r *Route) GetFullPath() string {
	if r.parent == nil {
		return r.path
	}
	return r.parent.GetFullPath() + r.path
}

func (r *Route) GET(h interface{}) *Operation {
	return r.GETNested(h, nil)
}

func (r *Route) GETNested(h interface{}, cb func(*Operation)) *Operation {
	o := GET(h)
	o.route = r
	r.operations.Append(o)

	if cb != nil {
		cb(o)
	}

	return o
}

func (r *Route) POST(h interface{}) *Operation {
	return r.POSTNested(h, nil)
}

func (r *Route) POSTNested(h interface{}, cb func(*Operation)) *Operation {
	o := POST(h)
	o.route = r
	r.operations.Append(o)

	if cb != nil {
		cb(o)
	}

	return o
}

func (r *Route) PUT(h interface{}) *Operation {
	return r.PUTNested(h, nil)
}

func (r *Route) PUTNested(h interface{}, cb func(*Operation)) *Operation {
	o := PUT(h)
	o.route = r
	r.operations.Append(o)

	if cb != nil {
		cb(o)
	}

	return o
}

func (r *Route) PATCH(h interface{}) *Operation {
	return r.PATCHNested(h, nil)
}

func (r *Route) PATCHNested(h interface{}, cb func(*Operation)) *Operation {
	o := PUT(h)
	o.route = r
	r.operations.Append(o)

	if cb != nil {
		cb(o)
	}

	return o
}

func (r *Route) DELETE(h interface{}) *Operation {
	return r.DELETENested(h, nil)
}

func (r *Route) DELETENested(h interface{}, cb func(*Operation)) *Operation {
	o := DELETE(h)
	o.route = r
	r.operations.Append(o)

	if cb != nil {
		cb(o)
	}

	return o
}

func (r *Route) PathParam(name string) *Param {
	p := PathParam(name)
	r.params.Append(p)
	return p
}

func (r *Route) QueryParam(name string) *Param {
	p := QueryParam(name)
	r.params.Append(p)
	return p
}

func (r *Route) HeaderParam(name string) *Param {
	p := HeaderParam(name)
	r.params.Append(p)
	return p
}

func (r *Route) BodyParam(name string) *Param {
	p := BodyParam(name)
	r.params.Append(p)
	return p
}

func (r *Route) Operations(ops ...*Operation) *Route {
	for _, o := range ops {
		o = o.Clone()
		o.route = r
		r.operations.Append(o)
	}
	return r
}

func (r *Route) Params(params ...*Param) *Route {
	r.params.Set(params...)
	return r
}

func (r *Route) Schemes(schemes ...string) *Route {
	r.schemes = schemes
	return r
}

func (r *Route) Accepts(accepts ...string) *Route {
	r.accepts = accepts
	return r
}

func (r *Route) Returns(returns ...string) *Route {
	r.returns = returns
	return r
}

func (r *Route) SchemesOnly(schemes ...string) *Route {
	r.schemes = schemes
	r.schemesOnly = true
	return r
}

func (r *Route) AcceptsOnly(accepts ...string) *Route {
	r.accepts = accepts
	r.acceptsOnly = true
	return r
}

func (r *Route) ReturnsOnly(returns ...string) *Route {
	r.returns = returns
	r.returnsOnly = true
	return r
}

func (r *Route) With(mw ...interface{}) *Route {
	for i := 0; i < len(mw); i++ {
		switch t := mw[i].(type) {
		case Middleware:
			r.middleware = append(r.middleware, t)
		case func(Handler) Handler:
			r.middleware = append(r.middleware, MiddlewareFunc(t))
		}
	}
	return r
}

func (r *Route) GetMiddleware() []Middleware {
	return r.middleware
}

func (r *Route) getAllMiddleware() {
	if r.parent != nil {
		r.middleware = middlewareUnion(r.middleware, r.parent.middleware)
	}
}

func (r *Route) GetParams() *Params {
	return r.params
}

func (r *Route) GetOperations() *Operations {
	return r.operations
}

func (r *Route) getAllParams() {
	if r.parent != nil {
		params := r.params.Clone()
		params.Union(r.parent.params)
		r.params = params
	}
}

func (r *Route) GetSchemes() []string {
	return r.schemes
}

func (r *Route) getAllSchemes() {
	if !r.schemesOnly {
		if r.parent != nil {
			r.schemes = stringsUnion(r.schemes, r.parent.schemes)
		}
	}
}

func (r *Route) GetAccepts() []string {
	return r.accepts
}

func (r *Route) getAllAccepts() {
	if !r.acceptsOnly {
		if r.parent != nil {
			r.accepts = stringsUnion(r.accepts, r.parent.accepts)
		}
	}
}

func (r *Route) GetReturns() []string {
	return r.returns
}

func (r *Route) getAllReturns() {
	if !r.returnsOnly {
		if r.parent != nil {
			r.returns = stringsUnion(r.returns, r.parent.returns)
		}
	}
}

func (r *Route) getAllPresets() {
	if r.parent != nil {
		r.presets.Union(r.parent.presets)
	}
}

func (r *Route) ServeHTTP(c *Context, res result) interface{} {
	c.route = r

	o := r.operations.GetByMethod(c.Request.Request.Method)

	err := setPathParams(c, o.params, res)
	if err != nil {
		return err
	}

	if o == nil {
		c.Response.encoding, err = getEncodingFromAccept(r.returns, c.Request)
		if err != nil {
			return err
		}
		mnaHandler := r.GetMethodNotAllowedHandler()
		if mnaHandler == nil {
			return errors.New(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		}
		return mnaHandler.ServeHTTP(c)
	}
	return o.ServeHTTP(c)
}

func (r *Route) Apply(p ...Preset) {
	r.presets = append(r.presets, p...)
}

func (r *Route) finalize() {
	r.getAllParams()
	r.getAllSchemes()
	r.getAllAccepts()
	r.getAllReturns()
	r.getAllMiddleware()
	r.getAllPresets()
	for _, cr := range r.node.getChildRoutes() {
		cr.finalize()
	}
	r.operations.finalize()
	r.params = nil
	r.schemes = nil
	r.accepts = nil
	r.returns = nil
	r.middleware = nil
	r.presets = nil
}

func setPathParams(c *Context, params *Params, res result) *errors.Error {
	var name string
	var value string
	var p *Param
	var pv *validation.Value
	if res.node.paramsCount > 0 {
		for i := 0; i < res.node.paramsCount; i++ {
			name, value = res.paramByIndex(i)
			if name == "" || value == "" {
				return nil //error
			}
			p = params.Get(name)
			if p == nil || !p.IsIn(IN_PATH) {
				return nil //error
			}
			pv = validation.NewValue(p.name, value, "path", p.GetAs())
			c.input.Append(pv)
		}
	}
	return nil

}
