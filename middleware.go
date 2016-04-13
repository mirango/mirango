package mirango

import (
	"fmt"
	"strings"

	"github.com/mirango/defaults"
	"github.com/mirango/framework"
	"github.com/mirango/validation"
)

type Middleware interface {
	Run(Handler) Handler
}

type MiddlewareFunc func(Handler) Handler

func (f MiddlewareFunc) Run(h Handler) Handler {
	return f(h)
}

func CheckSchemes(o *Operation) MiddlewareFunc {
	return MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c *Context) interface{} {

			var schemeAccepted bool
			if c.URL.Scheme == "" {
				c.URL.Scheme = framework.SCHEME_HTTP
			}

			schemeAccepted = containsString(o.GetAllSchemes(), c.URL.Scheme)
			if !schemeAccepted {
				return nil
			}

			return next.ServeHTTP(c)
		})
	})
}

func getEncodingFromAccept(returns []string, r *Request) (string, *Error) {
	var encoding string

	parts := strings.Split(r.Header.Get(framework.HEADER_Accept), ",")

	for _, acceptMime := range parts {
		mime := strings.Trim(strings.Split(acceptMime, ";")[0], " ")
		if 0 == len(mime) || mime == "*/*" {
			if len(returns) == 0 {
				encoding = defaults.MimeType
				break
			} else {
				encoding = returns[0]
				break
			}
		} else {
			if containsString(returns, mime) {
				encoding = mime
				break
			}
		}
	}

	if len(parts) == 0 {
		encoding = defaults.MimeType
	}

	if len(encoding) == 0 {
		// return encoding, NewError(406, "Request", "Encoding requested not valid.")
		encoding = defaults.MimeType
	}

	return encoding, nil
}

func CheckReturns(o *Operation) MiddlewareFunc {
	return MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c *Context) interface{} {
			// if mimeInAccept {
			enc, err := getEncodingFromAccept(o.GetAllReturns(), c.Request)
			if err != nil {
				return err
			}
			c.encoding = enc
			// } else {
			//
			// }

			return next.ServeHTTP(c)
		})
	})
}

func CheckAccepts(o *Operation) MiddlewareFunc {
	return MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c *Context) interface{} {
			// enc, err := getEncodingFromAccept(accepts, c.Request)
			// if err != nil {
			// 	return err
			// }
			// c.encoding = enc

			return next.ServeHTTP(c)
		})
	})
}

func CheckParams(o *Operation) MiddlewareFunc {
	return MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c *Context) interface{} {
			params := o.GetAllParams()
			var errs *validation.Error

			q := c.URL.Query()
			h := c.Request.Header

			if params.ContainsFiles() {
				c.ParseMultipartForm(defaults.MaxMemory)
			}
			if params.ContainsBodyParams() {
				c.ParseForm()
			}
			for _, p := range params.GetAll() {
				var pv *validation.Value
				if p.IsIn(IN_QUERY) {
					v, ok := q[p.name]
					if !ok {
						if p.IsRequired() {
							if errs == nil {
								errs = &validation.Error{}
							}
							errs.Append(p.name, fmt.Errorf("param %s is required", p.name))
						}
					} else {
						if !p.IsMultiple() {
							pv = validation.NewValue(p.name, v[0], "query", p.GetAs())
						} else {
							pv = validation.NewMultipleValue(p.name, v, "query", p.GetAs())
						}
					}
				} else if p.IsIn(IN_HEADER) {
					v, ok := h[p.name]
					if !ok {
						if p.IsRequired() {
							if errs == nil {
								errs = &validation.Error{}
							}
							errs.Append(p.name, fmt.Errorf("param %s is required", p.name))
						}
					} else {
						if !p.IsMultiple() {
							pv = validation.NewValue(p.name, v[0], "header", p.GetAs())
						} else {
							pv = validation.NewMultipleValue(p.name, v, "header", p.GetAs())
						}
					}
				} else if p.IsIn(IN_BODY) { // decide what to do when content type is form-encoded
					if p.IsFile() {
						_, ok := c.MultipartForm.File[p.name]
						if !ok {
							if p.IsRequired() {
								if errs == nil {
									errs = &validation.Error{}
								}
								errs.Append(p.name, fmt.Errorf("param %s is required", p.name))
							}
						} else {
							//pv = NewFileParamValue(p.name, v[0], "header")
						}
					} else if !p.IsFile() && params.ContainsFiles() {
						_, ok := c.MultipartForm.Value[p.name]
						if !ok {
							if p.IsRequired() {
								if errs == nil {
									errs = &validation.Error{}
								}
								errs.Append(p.name, fmt.Errorf("param %s is required", p.name))
							}
						} else {
							//pv = NewFileParamValue(p.name, v[0], "header")
						}
					} else {
						v, ok := c.Form[p.name]
						if !ok {
							if p.IsRequired() {
								if errs == nil {
									errs = &validation.Error{}
								}
								errs.Append(p.name, fmt.Errorf("param %s is required", p.name))
							}
						} else {
							if !p.IsMultiple() {
								pv = validation.NewValue(p.name, v[0], "body", p.GetAs())
							} else {
								pv = validation.NewMultipleValue(p.name, v, "body", p.GetAs())
							}
						}
					}
				}
				if pv != nil {
					c.Input[p.name] = pv
				}
			}

			vErrs := params.ValidateAll(c, c.Input)
			if vErrs != nil {
				if errs == nil {
					errs = &validation.Error{}
				}
				errs.UnionAppend(*vErrs)
			}

			if errs != nil {
				return errs
			}

			return next.ServeHTTP(c)
		})
	})
}
