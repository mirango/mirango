// Mirango is a conveniently smart web framework that is built with reusibility and easiness in mind.
//
// Mirango supports the following handler types:
//	mirango.Handler
//	mirango.HandlerFunc
//	func(*Context) interface{}
//	func(*Context)
//	framework.Handler
//	framework.HandlerFunc
//	func(framework.Context) interface{}
//	func(framework.Context)
//	func(*Response, *Request) interface{}
//	func(*Response, *Request)
//	func(framework.Response, framework.Request) interface{}
//	func(framework.Response, framework.Request)
//	http.Handler
//	http.HandlerFunc
//	func(http.ResponseWriter, *http.Request) interface{}
//	func(http.ResponseWriter, *http.Request)
//
package mirango

import (
	"net/http"

	"github.com/mirango/framework"
	"github.com/mirango/server"
)

type Mirango struct {
	*node
	*Route
	server        framework.Server
	renderers     framework.Renderers
	logger        framework.Logger
	sessionStores []framework.SessionStore // add ability to make sessions on different stores
}

func New(path interface{}) *Mirango {
	r := NewRoute(path)
	m := &Mirango{
		Route: r,
	}
	m.node = r.node.getRoot()
	r.mirango = m
	return m
}

func (m *Mirango) Renderers(r ...framework.Renderer) {
	m.renderers.Append(r...)
}

func (m *Mirango) Optimize() {
	// do optimization to routes and operations so that every object has what it needs
}

func (m *Mirango) Logger(l framework.Logger) {
	m.logger = l
}

func (m *Mirango) SessionStore(ss framework.SessionStore) {
	if ss != nil {
		m.sessionStores = append(m.sessionStores, ss)
	}
}

func (m *Mirango) Params(params ...*Param) *Mirango {
	m.Route.Params(params...)
	return m
}

func (m *Mirango) Path(path string) *Mirango {
	m.Route.Path(path)
	return m
}

func (m *Mirango) Server(s framework.Server) {
	m.server = s
}

func (m *Mirango) Prepare() {
	m.route.finalize()
	m.node.finalize()
}

func (m *Mirango) Start(addr string) error {
	m.Prepare()

	if m.server == nil {
		m.server = server.New()
	}
	m.server.SetHandler(m)
	m.server.SetLogger(m.logger)
	m.server.SetAddr(addr)
	return m.server.ListenAndServe()
}

func (m *Mirango) StartTLS(addr string, certFile string, keyFile string) error {
	m.Prepare()

	if m.server == nil {
		m.server = server.New()
	}
	m.server.SetHandler(m)
	m.server.SetLogger(m.logger)
	m.server.SetAddr(addr)
	return m.server.ListenAndServeTLS(certFile, keyFile)
}

func (m *Mirango) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nr := NewRequest(r)
	nw := NewResponse(w, m.renderers)
	c := NewContext(nw, nr)

	// defer log

	if m.logger != nil {
		c.LogWriter = m.logger.Logger(c)
	}

	if m.sessionStores != nil {
		var ses framework.Sessions
		var err error

		for _, ss := range m.sessionStores {
			ses, err = ss.GetAll(r)
			if err != nil {
				// log
				c.WriteHeader(http.StatusInternalServerError)
				return
			}
			c.sessions.Append(ses...)
		}
	}

	res, found, _ := m.node.match(r.URL.Path) // recommend redirection to standard path, tell if a match is found, otherwise return the latest found route
	if found && res.node != nil && res.node.route != nil {
		data := res.node.route.ServeHTTP(c, res)
		// check data type
		if !c.ended {
			err := c.sessions.Save(r, nw)
			if err != nil {
				// log
				c.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = nw.Render(c, data)
			if err != nil {
				// log
				c.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	} else {
		if res.node != nil {
			route := res.node.getRoute()
			if route != nil {
				h := route.GetNotFoundHandler()
				if h != nil {
					h.ServeHTTP(c)
					return
				}
			}
		}
		c.WriteHeader(http.StatusNotFound)
	}

	//log response
}
