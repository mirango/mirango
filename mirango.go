package mirango

import (
	"net/http"

	"github.com/mirango/framework"
	"github.com/mirango/server"
)

type Mirango struct {
	server        framework.Server
	route         *Route
	renderers     framework.Renderers
	logger        framework.Logger
	sessionStores []framework.SessionStore // add ability to make sessions on different stores
}

func New() *Mirango {
	r := NewRoute("")
	m := &Mirango{
		route: r,
	}
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
	m.route.params.Set(params...)
	return m
}

func (m *Mirango) GetParams() *Params {
	return m.route.params
}

func (m *Mirango) Route(path string) *Route {
	r := NewRoute(path)
	return m.AddRoute(r)
}

func (m *Mirango) AddRoute(r *Route) *Route {
	r = m.route.AddRoute(r)
	r.setMirango()
	return r
}

func (m *Mirango) Path(path string) *Mirango {
	m.route.SetPath(path)
	return m
}

func (m *Mirango) Server(s framework.Server) {
	m.server = s
}

func (m *Mirango) Start(addr string) error {
	if m.server == nil {
		m.server = server.New()
	}
	m.server.SetHandler(m)
	m.server.SetLogger(m.logger)
	m.server.SetAddr(addr)
	return m.server.ListenAndServe()
}

func (m *Mirango) StartTLS(addr string, certFile string, keyFile string) error {
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
			}
			c.sessions.AddMany(ses...)
		}
	}

	route, params := m.route.match(cleanPath(r.URL.Path)) // recommend redirection to standard path, tell if a match is found, otherwise return the latest found route
	if route != nil {
		data := route.ServeHTTP(c, params)
		// check data type
		if !c.ended {
			err := nw.Render(c, data)
			if err != nil {
				return
			}
		}
	} else {
	}
}
