package router

import (
	"ibanking-scraper/internal/authorizer"
	"ibanking-scraper/internal/http/middleware"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type (
	Router interface {
		Mount(e *echo.Echo, authorizer authorizer.Authorizer, middlewares ...echo.MiddlewareFunc)
	}

	Routers []Router
	Route   struct {
		Name        string
		Method      string
		Path        string
		Handler     echo.HandlerFunc
		Middlewares []echo.MiddlewareFunc

		Protected    bool
		AuthzOptions []authorizer.Option
	}

	Group struct {
		Prefix      string
		Middlewares []echo.MiddlewareFunc
		Routes      []*Route
		groups      []*Group
	}
)

func New(prefix string, middlewares ...echo.MiddlewareFunc) *Group {
	return &Group{
		Prefix:      prefix,
		Middlewares: middlewares,
	}
}

// RegisterFx register function with return value router interface
// to fx compatible under 'router' group
func RegisterFx(routerFunc interface{}) fx.Option {
	return fx.Provide(fx.Annotated{
		Group:  "router",
		Target: routerFunc,
	})
}

func (rs Routers) Mounts(e *echo.Echo, authorizer authorizer.Authorizer, middlewares ...echo.MiddlewareFunc) {
	for _, router := range rs {
		router.Mount(e, authorizer, middlewares...)
	}
}

func (r *Route) Slug(slugname string) *Route {
	r.Name = slugname

	return r
}

func (r *Route) Restricted(opts ...authorizer.Option) *Route {
	r.Protected = true
	r.AuthzOptions = append(r.AuthzOptions, opts...)
	return r
}

func (r *Route) Mount(e *echo.Echo, authorizer authorizer.Authorizer, middL ...echo.MiddlewareFunc) {
	if r.Protected {
		middL = append(middL, middleware.WithAuthorizer(authorizer, r.AuthzOptions...))
	}

	middL = append(middL, r.Middlewares...)
	echoroute := e.Add(r.Method, r.Path, r.Handler, middL...)

	if r.Name != "" {
		echoroute.Name = r.Name
	}

	r.Name = echoroute.Name
}

func (g *Group) build() []*Route {
	var routes []*Route

	for _, route := range g.Routes {
		route.Middlewares = append(g.Middlewares, route.Middlewares...)
		routes = append(routes, route)
	}

	for _, gr := range g.groups {
		routes = append(routes, gr.build()...)
	}

	return routes
}

func (g *Group) Mount(e *echo.Echo, authorizer authorizer.Authorizer, middL ...echo.MiddlewareFunc) {
	routes := g.build()

	for _, route := range routes {
		var midds []echo.MiddlewareFunc

		if route.Protected {
			midds = append(midds, middleware.WithAuthorizer(authorizer, route.AuthzOptions...))
		}

		midds = append(midds, middL...)
		midds = append(midds, route.Middlewares...)
		echoroute := e.Add(route.Method, route.Path, route.Handler, midds...)

		if route.Name != "" {
			echoroute.Name = route.Name
		}
		route.Name = echoroute.Name
	}
}

func (g *Group) Group(prefix string, middlewares ...echo.MiddlewareFunc) *Group {
	group := &Group{
		Prefix:      g.Prefix + prefix,
		Middlewares: append(g.Middlewares, middlewares...),
	}

	g.groups = append(g.groups, group)
	return group
}

func (g *Group) add(method, path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	route := &Route{
		Method:      method,
		Path:        g.Prefix + path,
		Handler:     handler,
		Middlewares: append(g.Middlewares, middlewares...),
	}

	g.Routes = append(g.Routes, route)
	return route
}

func (g *Group) CONNECT(path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	return g.add(http.MethodConnect, path, handler, middlewares...)
}

func (g *Group) HEAD(path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	return g.add(http.MethodHead, path, handler, middlewares...)
}

func (g *Group) GET(path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	return g.add(http.MethodGet, path, handler, middlewares...)
}

func (g *Group) POST(path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	return g.add(http.MethodPost, path, handler, middlewares...)
}

func (g *Group) PUT(path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	return g.add(http.MethodPut, path, handler, middlewares...)
}

func (g *Group) PATCH(path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	return g.add(http.MethodPatch, path, handler, middlewares...)
}

func (g *Group) DELETE(path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	return g.add(http.MethodDelete, path, handler, middlewares...)
}

func (g *Group) OPTIONS(path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	return g.add(http.MethodOptions, path, handler, middlewares...)
}

func (g *Group) TRACE(path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) *Route {
	return g.add(http.MethodTrace, path, handler, middlewares...)
}
