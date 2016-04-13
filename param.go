package mirango

import (
	"fmt"

	"github.com/mirango/framework"
	"github.com/mirango/validation"
)

type Params struct {
	params             map[string]*Param
	containsFiles      bool
	containsBodyParams bool
}

type paramIn int

const (
	IN_PATH paramIn = iota
	IN_QUERY
	IN_HEADER
	IN_BODY
)

func NewParams() *Params {
	return &Params{
		params: map[string]*Param{},
	}
}

func (p *Params) Get(name string) *Param {
	return p.params[name]
}

func (p *Params) GetAll() map[string]*Param {
	return p.params
}

func (p *Params) ContainsFiles() bool {
	return p.containsFiles
}

func (p *Params) ContainsBodyParams() bool {
	return p.containsBodyParams
}

func (p *Params) Append(params ...*Param) {
	for i := 0; i < len(params); i++ {
		name := params[i].name
		if name == "" {
			panic("Detected a param without a name.")
		}
		if _, ok := p.params[name]; ok {
			panic(fmt.Sprintf("Detected 2 params with the same name: \"%s\".", name))
		}
		// check if contains files or body params
		p.params[name] = params[i]
	}
}

func (p *Params) Union(params *Params) {
	for _, pa := range params.params {
		p.Append(pa)
	}
}

func (p *Params) Clone() *Params {
	params := NewParams()
	params.Union(p)
	return params
}

func (p *Params) Set(params ...*Param) {
	p.params = map[string]*Param{}
	p.Append(params...)
}

func (p *Params) Len() int {
	return len(p.params)
}

// Param
type Param struct {
	name          string
	validators    []validation.Validator
	def           interface{}
	as            framework.ValueType
	strSep        string
	isRequired    bool
	isMultiple    bool
	isFile        bool
	isInPath      bool
	isInQuery     bool
	isInHeader    bool
	isInBody      bool
	preprocessor  func(*validation.Value)
	postprocessor func(*validation.Value)
}

func NewParam(name string) *Param {
	return &Param{
		name: name,
		as:   framework.TYPE_STRING,
	}
}

func PathParam(name string) *Param {
	return NewParam(name).In(IN_PATH)
}

func QueryParam(name string) *Param {
	return NewParam(name).In(IN_QUERY)
}

func HeaderParam(name string) *Param {
	return NewParam(name).In(IN_HEADER)
}

func BodyParam(name string) *Param {
	return NewParam(name).In(IN_BODY)
}

func (p *Param) Name(name string) *Param {
	p.name = name
	return p
}

func (p *Param) GetName() string {
	return p.name
}

func (p *Param) Required() *Param {
	p.isRequired = true
	return p
}

func (p *Param) IsRequired() bool {
	return p.isRequired
}

func (p *Param) File() *Param {
	p.isFile = true
	return p
}

func (p *Param) IsFile() bool {
	return p.isFile
}

func (p *Param) Multiple() *Param {
	p.isMultiple = true
	return p
}

func (p *Param) IsMultiple() bool {
	return p.isMultiple
}

func (p *Param) As(as framework.ValueType) *Param {
	if as == framework.TYPE_STRING ||
		as == framework.TYPE_INT ||
		as == framework.TYPE_INT64 ||
		as == framework.TYPE_FLOAT ||
		as == framework.TYPE_FLOAT64 ||
		as == framework.TYPE_BOOL {

		p.as = as
	}
	return p
}

func (p *Param) GetAs() framework.ValueType {
	return p.as
}

// If a Param is in path then it is required.
func (p *Param) In(in ...paramIn) *Param {
	for _, i := range in {
		switch i {
		case IN_PATH:
			p.isInPath = true
			p.isRequired = true
		case IN_QUERY:
			p.isInQuery = true
		case IN_HEADER:
			p.isInHeader = true
		case IN_BODY:
			p.isInBody = true
		}
	}
	return p
}

// If a Param is in path then it is required.
func (p *Param) IsIn(in ...paramIn) bool {
	for _, i := range in {
		switch i {
		case IN_PATH:
			if !p.isInPath {
				return false
			}
		case IN_QUERY:
			if !p.isInQuery {
				return false
			}
		case IN_HEADER:
			if !p.isInHeader {
				return false
			}
		case IN_BODY:
			if !p.isInBody {
				return false
			}
		}
	}
	return true
}

// Must sets the validators to use.
func (p *Param) Validators(validators ...validation.Validator) *Param {
	p.validators = nil
	for _, v := range validators {
		if v != nil {
			p.validators = append(p.validators, v)
		}
	}
	return p
}

func (p *Param) validate(c framework.Context, v framework.ParamValue) error {
	for _, va := range p.validators {
		err := va.Validate(c, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Validate returns the first error it encountered
func (p *Param) Validate(c framework.Context, v framework.ParamValue) *validation.Error {
	err := p.validate(c, v)
	if err != nil {
		errs := &validation.Error{}
		errs.Append(p.name, err)
		return errs
	}
	return nil
}

func (p *Param) validateAll(c framework.Context, v framework.ParamValue) []error {
	var errs []error
	var err error
	for _, va := range p.validators {
		err = va.Validate(c, v)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// ValidateAll returns all the errors it encountered
func (p *Param) ValidateAll(c framework.Context, v framework.ParamValue) *validation.Error {
	var errs *validation.Error
	err := p.validateAll(c, v)
	if len(err) > 0 {
		if errs == nil {
			errs = &validation.Error{}
		}
		errs.Set(p.name, err)
	}
	return errs
}

// Validate returns the first error it encountered
func (pa *Params) Validate(c framework.Context, vs framework.ParamValues) *validation.Error {
	var err *validation.Error
	for _, p := range pa.params {
		err = p.Validate(c, vs[p.name])
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateFirst returns the first error it encountered for each param
func (pa *Params) ValidateFirst(c framework.Context, vs framework.ParamValues) *validation.Error {
	var errs *validation.Error
	var err error
nextParam:
	for _, p := range pa.params {
		err = p.validate(c, vs[p.name])
		if err != nil {
			if errs == nil {
				errs = &validation.Error{}
			}
			errs.Append(p.name, err)
			continue nextParam
		}
	}
	return errs
}

// ValidateAll returns all the errors it encountered for each param
func (pa *Params) ValidateAll(c framework.Context, vs framework.ParamValues) *validation.Error {
	var errs *validation.Error
	var err []error
	for _, p := range pa.params {
		err = p.validateAll(c, vs[p.name])
		if len(err) > 0 {
			if errs == nil {
				errs = &validation.Error{}
			}
			errs.Set(p.name, err)
		}
	}
	return errs
}
