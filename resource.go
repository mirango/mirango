package mirango

import (
	"strings"

	"github.com/gedex/inflector"
)

type resourceIndex interface {
	Index(c *Context) interface{}
}

type resourceGet interface {
	Get(c *Context) interface{}
}

type resourceCreate interface {
	Create(c *Context) interface{}
}

type resourceUpdate interface {
	Update(c *Context) interface{}
}

type resourcePatch interface {
	Patch(c *Context) interface{}
}

type resourceDelete interface {
	Delete(c *Context) interface{}
}

type Resource struct {
	Name        string
	Route       *Route
	EntityRoute *Route
	Index       *Operation
	Create      *Operation
	Get         *Operation
	Update      *Operation
	Patch       *Operation
	Delete      *Operation
	re          interface{}
}

func NewResource(name string, re interface{}) *Resource {
	return NewResourceNested(name, re, nil)
}

func NewResourceNested(name string, re interface{}, cb func(*Resource)) *Resource {
	if re == nil {
		panic("resource handler is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		panic("invalid resource name")
	}

	name = strings.ToLower(name)

	route := NewRoute(name)
	paramName := inflector.Singularize(name) + "_id"

	eRoute := route.Branch(":" + paramName)
	eRoute.PathParam(paramName)

	foundOne := false

	var index, create, get, update, patch, delete *Operation

	if t, ok := re.(resourceIndex); ok {
		foundOne = true
		index = route.GET(t.Index).Name("get_" + inflector.Pluralize(name))
	}

	if t, ok := re.(resourceCreate); ok {
		foundOne = true
		create = route.POST(t.Create).Name("create_" + inflector.Singularize(name))
	}

	if t, ok := re.(resourceGet); ok {
		foundOne = true
		get = eRoute.GET(t.Get).Name("get_" + inflector.Singularize(name))
	}

	if t, ok := re.(resourceUpdate); ok {
		foundOne = true
		update = eRoute.PUT(t.Update).Name("update_" + inflector.Singularize(name))
	}

	if t, ok := re.(resourcePatch); ok {
		foundOne = true
		patch = eRoute.PATCH(t.Patch).Name("patch_" + inflector.Singularize(name))
	}

	if t, ok := re.(resourceDelete); ok {
		foundOne = true
		delete = eRoute.DELETE(t.Delete).Name("delete_" + inflector.Singularize(name))
	}

	if !foundOne {
		panic("resource handler should contain at least one method")
	}

	resource := &Resource{
		Name:        name,
		Route:       route,
		EntityRoute: eRoute,
		Index:       index,
		Create:      create,
		Get:         get,
		Update:      update,
		Patch:       patch,
		Delete:      delete,
		re:          re,
	}

	if cb != nil {
		cb(resource)
	}

	return resource
}

func (r *Resource) Resource(name string, re interface{}) *Resource {
	return r.ResourceNested(name, re, nil)
}

func (r *Resource) ResourceNested(name string, re interface{}, cb func(*Resource)) *Resource {
	resource := NewResource(name, re)
	resource = r.EntityRoute.AddResource(resource)

	if cb != nil {
		cb(resource)
	}

	return resource
}

func (r *Resource) AddResource(resource *Resource) *Resource {
	return r.AddResourceNested(resource, nil)
}

func (r *Resource) AddResourceNested(resource *Resource, cb func(*Resource)) *Resource {
	if resource == nil {
		panic("resource is nil")
	}

	resource = resource.Clone()

	resource.Route = r.Route.AddRoute(resource.Route)
	resource.EntityRoute = resource.Route.AddRoute(resource.EntityRoute)

	if cb != nil {
		cb(resource)
	}

	return resource
}

func (r *Resource) Clone() *Resource {
	resource := NewResource(r.Name, r.re)
	return resource
}

func (r *Route) Resource(name string, re interface{}) *Resource {
	return r.ResourceNested(name, re, nil)
}

func (r *Route) ResourceNested(name string, re interface{}, cb func(*Resource)) *Resource {
	resource := NewResource(name, re)
	resource = r.AddResource(resource)

	if cb != nil {
		cb(resource)
	}

	return resource
}

func (r *Route) AddResource(resource *Resource) *Resource {
	return r.AddResourceNested(resource, nil)
}

func (r *Route) AddResourceNested(resource *Resource, cb func(*Resource)) *Resource {
	if resource == nil {
		panic("resource is nil")
	}

	resource = resource.Clone()

	resource.Route = r.AddRoute(resource.Route)
	resource.EntityRoute = resource.Route.AddRoute(resource.EntityRoute)

	if cb != nil {
		cb(resource)
	}

	return resource
}
