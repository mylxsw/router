package web

import (
	"fmt"
	"net"
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

	decorators           []HandlerDecorator
	exceptionHandler     ExceptionHandler
	routeNotFoundHandler RouteNotFoundHandler
	logger               Log
}

// ExceptionHandler is a function interface for exception handler
type ExceptionHandler func(wtx Context, err error) Response

// RouteNotFoundHandler ias a function interface for route not found handler
type RouteNotFoundHandler func(wtx Context, route RealRoute) Response

// NewRouter create a new Router
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

// WithExceptionHandler set a exception handler function
func (router *Router) WithExceptionHandler(fn ExceptionHandler) *Router {
	router.exceptionHandler = fn
	return router
}

// WithRouteNotFoundHandler set a route not found handler function
func (router *Router) WithRouteNotFoundHandler(fn RouteNotFoundHandler) *Router {
	router.routeNotFoundHandler = fn
	return router
}

// WithLogger set a logger for router
func (router *Router) WithLogger(logger Log) *Router {
	router.logger = logger
	return router
}

// Serve accepts incoming connections on the Listener l
func (router *Router) Serve(l net.Listener) error {
	return http.Serve(l, router)
}

// ServeTLS accepts incoming connections on the Listener l
func (router *Router) ServeTLS(l net.Listener, certFile, keyFile string) error {
	return http.ServeTLS(l, router, certFile, keyFile)
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve with router to handle requests on incoming connections.
func (router *Router) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, router)
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it
// expects HTTPS connections.
func (router *Router) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, router)
}

// Group create a router group
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

// Routes return all routes as a slice
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

func (router *Router) handleRouteNotFound(wtx Context, route RealRoute) {
	if router.routeNotFoundHandler == nil {
		_ = wtx.HTMLWithCode("Not Found", http.StatusNotFound).Send()
		return
	}

	if resp := router.routeNotFoundHandler(wtx, route); resp != nil {
		_ = resp.Send()
	}
}

func (router *Router) handleException(wtx Context, err error) Response {
	if router.exceptionHandler == nil {
		return NewErrorResponse(wtx.Response(), err.Error(), http.StatusInternalServerError)
	}

	return router.exceptionHandler(wtx, err)
}

func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	realRoute := NewRealRoute(request)
	matchedRoute, pathVars := router.Match(realRoute)
	if matchedRoute == nil {
		router.handleRouteNotFound(NewWebContext(router, nil, writer, request), realRoute)
		return
	}

	ctx := NewWebContext(router, pathVars, writer, request)
	router.handle(ctx, matchedRoute)
}

func (router *Router) handle(ctx Context, matchedRoute Route) {

	handler := func(ctx Context) (resp Response) {
		ctxCB := func() Context { return ctx }
		reqCB := func() Request { return ctx.Request() }
		respCB := func() Responsor { return ctx.Response() }
		matchedRouteCB := func() Route { return matchedRoute }

		provider, _ := router.cc.Provider(ctxCB, reqCB, respCB, matchedRouteCB)

		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case error:
					resp = router.handleException(ctx, err.(error))
				default:
					resp = router.handleException(ctx, fmt.Errorf("%v", err))
				}
			}
		}()

		results, err := router.cc.CallWithProvider(matchedRoute.Handle(), provider)
		if err != nil {
			return router.handleException(ctx, err)
		}

		return router.parseResponse(ctx, results)
	}

	decors := matchedRoute.Decorators()
	for i := range decors {
		d := decors[i]
		handler = d(handler)
	}

	if err := handler(ctx).Send(); err != nil && router.logger != nil {
		router.logger.Errorf("send response failed: %v", err)
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
		return ctx.HTML(results[0].(string))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return ctx.HTML(fmt.Sprintf("%d", results[0]))
	case float32, float64:
		return ctx.HTML(fmt.Sprintf("%f", results[0]))
	case error:
		if results[0] == nil {
			return ctx.HTML("")
		}

		panic(results[0])
	default:
		if jsonAble, ok := results[0].(JSONAble); ok {
			return ctx.JSON(jsonAble.ToJSON())
		}

		return ctx.JSON(results[0])
	}
}
