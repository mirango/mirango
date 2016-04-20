package mirango

import (
	"net/http"
	"strings"

	"github.com/mirango/errors"
	"github.com/mirango/validation"
)

type Route struct {
	mirango                 *Mirango
	parent                  *Route
	children                Routes
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

	slices           []string
	paramIndices     map[int]int
	paramNames       []string
	containsWildCard bool
}

type Routes []*Route

func (r Routes) Len() int {
	return len(r)
}

func (r Routes) Swap(i int, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Routes) Less(i int, j int) bool {
	if r[i].containsWildCard && !r[j].containsWildCard {
		return false
	}
	if !r[i].containsWildCard && r[j].containsWildCard {
		return true
	}
	if r[i].containsWildCard && r[j].containsWildCard {
		// check position of wildcard
	}
	return false
}

func NewRoute(path string) *Route {
	return &Route{
		path:         cleanPath(path),
		operations:   NewOperations(),
		params:       NewParams(),
		paramIndices: map[int]int{},
	}
}

type pathParam struct {
	Key   string
	Value string
}

type pathParams []*pathParam

func (r *Route) GetRoot() *Route {
	if r.parent != nil {
		return r.parent.GetRoot()
	}
	return r
}

func (r *Route) processPath() {
	//path := r.path
	r.paramNames = nil
	r.paramIndices = map[int]int{}
	slices := strings.Split(r.path[1:], "/")

	if len(slices) == 0 {
		panic("path is empty")
	}

	r.slices = slices

	// check that every var name has length more than 0

	for i, s := range slices {
		param := strings.LastIndex(s, ":")
		wildcardParam := strings.LastIndex(s, "*")
		if param > wildcardParam {
			r.paramNames = append(r.paramNames, s[param+1:])
			r.paramIndices[i] = param
		} else if param < wildcardParam && i == len(slices)-1 {
			r.paramNames = append(r.paramNames, s[wildcardParam+1:])
			r.paramIndices[i] = wildcardParam
			r.containsWildCard = true
			r.children = nil
		} else if param == -1 && wildcardParam == -1 {
			r.paramIndices[i] = -1
			continue
		}
	}

	// check all paths
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

func (r *Route) match(path string) (nr *Route, p pathParams) {
	route := r

	params := 0

	if r.parent == nil {
		nPath := r.path
		if nPath == path {
			nr = r
			p = nil
			return
		} else if !strings.HasPrefix(path, nPath) || len(nPath) > len(path) {
			nr = nil
			p = nil
			return
		}
		path = path[len(nPath):]
	}

	slices := strings.Split(path[1:], "/")

	param := -1
	wildcardParam := -1

look:
	for _, c := range route.children {
	walk:
		for j := range c.slices {
			param = -1
			wildcardParam = -1
			if c.containsWildCard {
				if j == len(c.slices)-1 {
					wildcardParam = c.paramIndices[j]
				}
			} else {
				param = c.paramIndices[j]
			}

			if param == -1 && wildcardParam == -1 {
				if c.slices[j] != slices[j] {
					if j == len(slices)-1 && j == len(c.slices)-1 {
						nr = nil
						p = nil
						return
					}
					continue look
				} else {
					if j == len(slices)-1 && j == len(c.slices)-1 {
						nr = c
						return
					} else if j == len(c.slices)-1 {
						route = c
						slices = slices[j+1:]
						params = 0
						goto look
					} else if j == len(slices)-1 {
						nr = nil
						p = nil
						return
					}
					continue walk
				}
			} else if (param > wildcardParam && (len(slices[j]) <= param || slices[j][:param] != c.slices[j][:param])) ||
				((param < wildcardParam) && (!c.containsWildCard || len(slices[j]) <= wildcardParam || slices[j][:wildcardParam] != c.slices[j][:wildcardParam])) {
				continue look
			} else if param > wildcardParam {
				p = append(p, &pathParam{c.paramNames[params], slices[j][param:]})
				params++
				if j == len(slices)-1 && j == len(c.slices)-1 {
					nr = c
					return
				} else if j == len(c.slices)-1 {
					route = c
					slices = slices[j+1:]
					params = 0
					goto look
				} else if j == len(slices)-1 {
					nr = nil
					p = nil
					return
				}
				continue walk
			} else if param < wildcardParam {
				p = append(p, &pathParam{c.paramNames[params], strings.TrimSuffix(slices[j][wildcardParam:]+"/"+strings.Join(slices[j+1:], "/"), "/")})
				nr = c
				return
			}
		}
	}
	return
}

func (r *Route) BuildPath(v ...interface{}) string {
	return ""
}

func (r *Route) setMirango() {
	rr := r.GetRoot()
	if rr != nil {
		r.mirango = rr.mirango
	}
}

func (r *Route) Sort() {

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

func (r *Route) Path(path string) {
	r.path = cleanPath(path)
}

func (r *Route) Branch(path string) *Route {
	nr := NewRoute(path)
	return r.AddRoute(nr)
}

func (r *Route) AddRoute(nr *Route) *Route {
	if nr == nil {
		panic("route is nil")
	}

	if nr.parent != nil {
		nr = nr.Clone()
	}

	if r.containsWildCard {
		panic("wildcard routes can not have sub-routes")
	}

	nr.parent = r

	nr.processPath()

	// check path

	r.children = append(r.children, nr)

	nr.setMirango()
	return nr
}

func (r *Route) GET(h interface{}) *Operation {
	o := GET(r, h)
	r.operations.Append(o)
	return o
}

func (r *Route) POST(h interface{}) *Operation {
	o := POST(r, h)
	r.operations.Append(o)
	return o
}

func (r *Route) PUT(h interface{}) *Operation {
	o := PUT(r, h)
	r.operations.Append(o)
	return o
}

func (r *Route) DELETE(h interface{}) *Operation {
	o := DELETE(r, h)
	r.operations.Append(o)
	return o
}

func (r *Route) Operations(ops ...*Operation) *Route {
	for _, o := range ops {
		o = o.Clone()
		o.route = r
		r.operations.Append(o)
	}
	return r
}

// Copy returns a pointer to a copy of the route.
// It does not copy parent, operations, nor deep-copy the params.
func (r *Route) Clone() *Route {
	route := NewRoute(r.path)
	// for _, cr := range rs {
	// 	route.AddRoute(cr.Copy())
	// }
	// route.path = r.path
	// route.operations = r.operations
	// route.params = r.params
	// route.middleware = r.middleware
	return route
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

func (r *Route) With(mw ...interface{}) *Route {
	for i := 0; i < len(mw); i++ {
		switch t := mw[i].(type) {
		case Middleware:
			r.middleware = append(r.middleware, t)
		case MiddlewareFunc:
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

func (r *Route) GetAllMiddleware() []Middleware {
	if r.parent != nil {
		return middlewareUnion(r.middleware, r.parent.GetAllMiddleware())
	}
	return r.middleware
}

func (r *Route) GetParams() *Params {
	return r.params
}

func (r *Route) GetAllParams() *Params {
	params := r.params.Clone()
	if r.parent != nil {
		params.Union(r.parent.GetAllParams())
	}
	return params
}

func (r *Route) GetSchemes() []string {
	return r.schemes
}

func (r *Route) GetAllSchemes() []string {
	if r.parent != nil {
		return stringsUnion(r.schemes, r.parent.GetAllSchemes())
	}
	return r.schemes
}

func (r *Route) GetAccepts() []string {
	return r.accepts
}

func (r *Route) GetAllAccepts() []string {
	if r.parent != nil {
		return stringsUnion(r.accepts, r.parent.GetAllAccepts())
	}
	return r.accepts
}

func (r *Route) GetReturns() []string {
	return r.returns
}

func (r *Route) GetAllReturns() []string {
	if r.parent != nil {
		return stringsUnion(r.returns, r.parent.GetAllReturns())
	}
	return r.returns
}

func (r *Route) ServeHTTP(c *Context, params pathParams) interface{} {
	c.route = r
	err := setPathParams(c, r.GetAllParams(), params)
	if err != nil {
		return err
	}

	o := r.operations.GetByMethod(c.Request.Request.Method)
	if o == nil {
		c.Response.encoding, err = getEncodingFromAccept(r.GetAllReturns(), c.Request)
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

func setPathParams(c *Context, params *Params, pathParams pathParams) *errors.Error {
	for _, p := range params.GetAll() {
		var pv *validation.Value
		if p.IsIn(IN_PATH) {
			var v string
			for i, par := range pathParams {
				if par.Key == p.name {
					v = par.Value
					pathParams = append(pathParams[:i], pathParams[i+1:]...)
					break
				}
			}
			if v == "" {
				return nil //error
			}
			pv = validation.NewValue(p.name, v, "path", p.GetAs())
			c.Input[p.name] = pv
		}
	}
	return nil
}

func cleanPath(path string) string {
	path = strings.ToLower(path)
	path = strings.Trim(path, "/")
	slices := strings.Split(path, "/")
	nPath := ""
	for _, s := range slices {
		if len(s) > 0 {
			nPath = nPath + "/" + s
		}
	}
	return nPath

}
