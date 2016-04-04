package mirango

import (
	"net/http"

	"github.com/wlMalk/mirango/framework"
	"github.com/wlMalk/mirango/internal/tree"
	"github.com/wlMalk/mirango/server"
)

type Mirango struct {
	server        framework.Server
	node          *tree.Node
	renderers     framework.Renderers
	logger        framework.Logger
	sessionStores []framework.SessionStore // add ability to make sessions on different stores
}

func New() *Mirango {
	r := NewRoute("")
	m := &Mirango{
		node:       tree.New(r),
	}
	r.node = m.node
	r.mirango = m
	return m
}

func (m *Mirango) Renderer(r framework.Renderer) {
	m.renderers.Add(r)
}

func (m *Mirango) Logger(l framework.Logger) {
	m.logger = l
}

func (m *Mirango) SessionStore(ss framework.SessionStore) {
	if ss != nil {
		m.sessionStores = append(m.sessionStores, ss)
	}
}

