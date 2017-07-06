package mirango

import (
	"net/http"

	"github.com/mirango/framework"
)

type Request struct {
	*http.Request
	input    framework.ParamValues
	sessions framework.Sessions
}

func NewRequest(r *http.Request) *Request {
	return &Request{
		Request: r,
	}
}

// Param returns the input parameter value by its name.
func (r *Request) Param(name string) framework.ParamValue {
	return r.input.Get(name)
}

// ParamOk returns the input parameter value by its name.
func (r *Request) IsSet(name string) bool {
	p := r.input.Get(name)
	return p != nil
}

func (r *Request) Params(names ...string) framework.ParamValues {
	if len(names) == 0 {
		return r.input
	}
	var params framework.ParamValues
	for _, n := range names {
		p := r.input.Get(n)
		if p == nil {
			continue
		}
		params.Append(p)
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

func (r *Request) Session(name string) framework.Session {
	return r.sessions.Get(name)
}

func (r *Request) SetSessionValue(string, interface{}, interface{}) error {
	return nil
}

func (r *Request) GetSessionValue(string, interface{}) (framework.Value, error) {
	return nil, nil
}
