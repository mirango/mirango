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

type resourceDelete interface {
	Delete(c *Context) interface{}
}

type Resource struct {
	Route       *Route
	EntityRoute *Route
}

func NewResource(name string, path interface{}, re interface{}) *Resource {
	if re == nil {
		panic("resource handler is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		panic("invalid resource name")
	}

	paramName := inflector.Singularize(name) + "_id"

	route := NewRoute(path)
	eRoute := route.Branch(":" + paramName)

	eRoute.PathParam(paramName)

	foundOne := false

	if t, ok := re.(resourceIndex); ok {
		foundOne = true
		route.GET(t.Index).Name("get_" + inflector.Pluralize(name))
	}

	if t, ok := re.(resourceCreate); ok {
		foundOne = true
		route.POST(t.Create).Name("create_" + inflector.Singularize(name))
	}

	if t, ok := re.(resourceGet); ok {
		foundOne = true
		eRoute.GET(t.Get).Name("get_" + inflector.Singularize(name))
	}

	if t, ok := re.(resourceUpdate); ok {
		foundOne = true
		eRoute.PUT(t.Update).Name("update_" + inflector.Singularize(name))
	}

	if t, ok := re.(resourceDelete); ok {
		foundOne = true
		eRoute.DELETE(t.Delete).Name("delete_" + inflector.Singularize(name))
	}

	if !foundOne {
		panic("resource handler should contain at least one method")
	}

	resource := &Resource{
		Route:       route,
		EntityRoute: eRoute,
	}

	return resource
}

func (r *Route) Resource(name string, path interface{}, re interface{}) *Resource {
	resource := NewResource(name, path, re)
	r.AddResource(resource)
	return resource
}

func (r *Route) AddResource(resource *Resource) *Resource {
	if resource == nil {
		panic("resource is nil")
	}

	nResource := &Resource{}

	if resource.Route.parent != nil {
		nResource.Route = r.AddRoute(resource.Route)
		nResource.EntityRoute = resource.Route.AddRoute(resource.EntityRoute)
	} else {
		nResource.Route = r.AddRoute(resource.Route)
	}

	return nResource
}
