package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/mylxsw/container"
)

// Router is route manager
type Router struct {
	lock           sync.RWMutex
	cc             container.Container
	routes         []Route
	routesByMethod map[string][]Route
	conf           *Config

	decorators []HandlerDecorator
}

func NewRouter(cc container.Container, conf *Config, decors ...HandlerDecorator) *Router {
	ccc := container.Extend(cc)
	ccc.MustSingleton(func() Decoder {
		return structDecoder{}
	})
	ccc.MustSingleton(func() *Config { return conf })

	return createRouter(ccc, conf, decors...)
}

func createRouter(cc container.Container, conf *Config, decors ...HandlerDecorator) *Router {
	return &Router{
		cc:             cc,
		routes:         make([]Route, 0),
		routesByMethod: make(map[string][]Route),
		conf:           conf,
		decorators:     decors,
	}
}

func (router *Router) Group(prefix string, f func(router *Router), decors ...HandlerDecorator) {
	groupRouter := createRouter(router.cc, router.conf, decors...)
	f(groupRouter)

	prefix = strings.Trim(prefix, "/")
	for _, r := range groupRouter.routes {
		route := NewRoute()
		route.WithMethod(r.Methods()...)
		route.WithHost(r.Hosts()...)
		route.WithContentTypes(r.ContentTypes()...)
		route.WithPath(prefix + "/" + strings.TrimLeft(r.Path(), "/"))
		route.WithHandler(r.Handle())
		route.WithDecorators(r.Decorators()...)
		router.AddRoute(route)
	}
}

func (router *Router) Get(pattern string, handler interface{}) Route {
	return router.Add([]string{"GET"}, pattern, handler)
}

func (router *Router) Post(pattern string, handler interface{}) Route {
	return router.Add([]string{"POST"}, pattern, handler)
}

func (router *Router) Put(pattern string, handler interface{}) Route {
	return router.Add([]string{"PUT"}, pattern, handler)
}

func (router *Router) Delete(pattern string, handler interface{}) Route {
	return router.Add([]string{"DELETE"}, pattern, handler)
}

func (router *Router) Patch(pattern string, handler interface{}) Route {
	return router.Add([]string{"PATCH"}, pattern, handler)
}

func (router *Router) Head(pattern string, handler interface{}) Route {
	return router.Add([]string{"HEAD"}, pattern, handler)
}

func (router *Router) Options(pattern string, handler interface{}) Route {
	return router.Add([]string{"OPTIONS"}, pattern, handler)
}

func (router *Router) Any(pattern string, handler interface{}) Route {
	return router.Add([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}, pattern, handler)
}

func (router *Router) Add(methods []string, pattern string, handler interface{}) Route {
	route := NewRoute()
	route.WithMethod(methods...)
	route.WithPath(pattern)
	route.WithHandler(handler)

	router.AddRoute(route)
	return route
}

// AddRoute add a SimpleRoute to router
func (router *Router) AddRoute(route Route) {
	if len(route.Methods()) == 0 {
		panic("Request method is required")
	}

	route.PrependDecorators(router.decorators...)

	router.lock.Lock()
	defer router.lock.Unlock()

	router.routes = append(router.routes, route)

	for _, m := range route.Methods() {
		if router.routesByMethod[m] == nil {
			router.routesByMethod[m] = make([]Route, 0)
		}

		router.routesByMethod[m] = append(router.routesByMethod[m], route)
	}
}

func (router *Router) Routes() []Route {
	router.lock.RLock()
	defer router.lock.RUnlock()

	return router.routes
}

// Match return whether current SimpleRoute is matched with registered routes
func (router *Router) Match(current RealRoute) (Route, map[string]string) {
	router.lock.RLock()
	defer router.lock.RUnlock()

	if rs, ok := router.routesByMethod[current.Method]; ok {
		for _, r := range rs {
			if matched, pathVars := r.Match(current); matched {
				return r, pathVars
			}
		}
	}

	return nil, nil
}

func (router *Router) notFoundHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	_, _ = writer.Write([]byte("Not Found"))
}

func (router *Router) exceptionHandler(ctx Context, err error) Response {
	return NewErrorResponse(ctx.Response(), err.Error(), http.StatusInternalServerError)
}

func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	matchedRoute, pathVars := router.Match(NewRealRoute(request))
	if matchedRoute == nil {
		router.notFoundHandler(writer, request)
		return
	}

	ctx := NewWebContext(router, pathVars, writer, request)
	router.handle(ctx, matchedRoute)
}

func (router *Router) handle(ctx Context, matchedRoute Route) {

	handler := func(ctx Context) Response {
		ctxCB := func() Context { return ctx }
		reqCB := func() Request { return ctx.Request() }
		respCB := func() Responsor { return ctx.Response() }
		matchedRouteCB := func() Route { return matchedRoute }

		provider, _ := router.cc.Provider(ctxCB, reqCB, respCB, matchedRouteCB)
		results, err := router.cc.CallWithProvider(matchedRoute.Handle(), provider)
		if err != nil {
			return router.exceptionHandler(ctx, err)
		}

		return router.parseResponse(ctx, results)
	}

	decors := matchedRoute.Decorators()
	for i := range decors {
		d := decors[i]
		handler = d(handler)
	}

	if err := handler(ctx).Send(); err != nil {
		log.Printf("send response failed: %v", err)
	}
}

func (router *Router) parseResponse(ctx Context, results []interface{}) Response {
	if len(results) == 0 {
		return ctx.Nil()
	}

	if len(results) > 1 {
		if err, ok := results[1].(error); ok {
			if err != nil {
				return router.exceptionHandler(ctx, err)
			}
		}
	}

	switch results[0].(type) {
	case Response:
		return results[0].(Response)
	case string:
		return ctx.NewHTMLResponse(results[0].(string))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return ctx.NewHTMLResponse(fmt.Sprintf("%d", results[0]))
	case float32, float64:
		return ctx.NewHTMLResponse(fmt.Sprintf("%f", results[0]))
	case error:
		if results[0] == nil {
			return ctx.HTML("")
		}

		panic(results[0])
	default:
		if jsonAble, ok := results[0].(JSONAble); ok {
			return ctx.NewJSONResponse(jsonAble.ToJSON())
		}

		return ctx.NewJSONResponse(results[0])
	}
}
