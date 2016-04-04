package mirango

import (
	"strings"

	"github.com/wlMalk/mirango/defaults"
	"github.com/wlMalk/mirango/framework"
	"github.com/wlMalk/mirango/internal/util"
	"github.com/wlMalk/mirango/validation"
)

type Middleware interface {
	Run(Handler) Handler
}

type MiddlewareFunc func(Handler) Handler

func (f MiddlewareFunc) Run(h Handler) Handler {
	return f(h)
}

