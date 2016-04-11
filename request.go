package mirango

import (
	"net/http"

	"github.com/wlMalk/mirango/framework"
)

type Request struct {
	*http.Request
	Input    framework.ParamValues
	route    *Route
	sessions framework.Sessions
}

func NewRequest(r *http.Request) *Request {
	return &Request{
		Request: r,
		Input:   framework.ParamValues{},
	}
}

// Param returns the input parameter value by its name.
func (r *Request) Param(name string) framework.ParamValue {
	return r.Input[name]
}

// ParamOk returns the input parameter value by its name.
func (r *Request) ParamOk(name string) (framework.ParamValue, bool) {
	p, ok := r.Input[name]
	return p, ok
}

func (r *Request) Params(names ...string) framework.ParamValues {
	if len(names) == 0 {
		return r.Input
	}
	params := framework.ParamValues{}
	for _, n := range names {
		p, ok := r.Input[n]
		if !ok {
			continue
		}
		params[n] = p
	}
	return params
}

func (r *Request) Path() string {
	return r.RequestURI
}
func (r *Request) Method() string {
	return r.Request.Method
}

func (r *Request) Sessions() framework.Sessions {
	return r.sessions
}

func (r *Request) Session(name string) (framework.Session, error) {
	return r.sessions.Get(name)
}

func (r *Request) SetSessionValue(string, interface{}, interface{}) error {
	return nil
}

func (r *Request) GetSessionValue(string, interface{}) (framework.Value, error) {
	return nil, nil
}
