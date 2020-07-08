package web

import (
	"fmt"
	"net/http"
	"strings"
)

// SimpleRoute is a route for request
type SimpleRoute struct {
	hosts        []string
	path         string
	methods      []string
	contentTypes []string
	handler      interface{}

	parsedPaths map[ParsedPathType][]ParsedPath
	decorators  []HandlerDecorator
}

func (route *SimpleRoute) WithDecorators(decors ...HandlerDecorator) {
	route.decorators = append(route.decorators, decors...)
}

func (route *SimpleRoute) PrependDecorators(decors ...HandlerDecorator) {
	route.decorators = append(decors, route.decorators...)
}

func (route *SimpleRoute) Decorators() []HandlerDecorator {
	return route.decorators
}

func (route *SimpleRoute) Hosts() []string {
	return route.hosts
}

func (route *SimpleRoute) Path() string {
	return route.path
}

func (route *SimpleRoute) ContentTypes() []string {
	return route.contentTypes
}

type ParsedPathType string

const (
	ParsedPathPlain       ParsedPathType = "plain"
	ParsedPathPlaceholder ParsedPathType = "placeholder"
)

type ParsedPath struct {
	Index   int
	Segment string
	Type    ParsedPathType
}

type RealRoute struct {
	Host         string
	Path         string
	PathSegments []string
	Method       string
	ContentType  string
}

// NewRealRoute create a new SimpleRoute from request
func NewRealRoute(request *http.Request) RealRoute {
	return RealRoute{
		Host:         strings.ToLower(strings.SplitN(request.Host, ":", 2)[0]),
		Method:       request.Method,
		Path:         request.URL.Path,
		PathSegments: pathSegments(request.URL.Path),
		ContentType:  strings.ToLower(request.Header.Get("Content-Type")),
	}
}

func pathSegments(path string) []string {
	segments := make([]string, 0)
	for _, s := range strings.Split(strings.Trim(path, "/"), "/") {
		if s == "" {
			continue
		}

		segments = append(segments, s)
	}

	return segments
}

func NewRoute() Route {
	return &SimpleRoute{
		hosts:        make([]string, 0),
		path:         "",
		methods:      make([]string, 0),
		contentTypes: make([]string, 0),
		handler:      nil,
		decorators:   make([]HandlerDecorator, 0),
	}
}

func (route *SimpleRoute) WithHandler(handler interface{}) {
	route.handler = handler
}

func (route *SimpleRoute) Methods() []string {
	return route.methods
}

func (route *SimpleRoute) WithHost(hosts ...string) {
	for _, h := range hosts {
		route.hosts = append(route.hosts, strings.ToLower(h))
	}
}

func (route *SimpleRoute) WithPath(path string) {
	route.path = strings.Trim(strings.ToLower(path), "/")

	route.parsedPaths = make(map[ParsedPathType][]ParsedPath)
	route.parsedPaths[ParsedPathPlain] = make([]ParsedPath, 0)
	route.parsedPaths[ParsedPathPlaceholder] = make([]ParsedPath, 0)

	index := 0
	for _, segment := range strings.Split(route.path, "/") {
		segment = strings.Trim(segment, " ")
		if segment == "" {
			continue
		}

		var segmentType ParsedPathType
		if len(segment) > 2 && segment[0] == '{' && segment[len(segment)-1] == '}' {
			segmentType = ParsedPathPlaceholder
			segment = segment[1 : len(segment)-1]
		} else {
			segmentType = ParsedPathPlain
		}

		route.parsedPaths[segmentType] = append(route.parsedPaths[segmentType], ParsedPath{
			Index:   index,
			Segment: segment,
			Type:    segmentType,
		})

		index++
	}
}

func (route *SimpleRoute) WithMethod(methods ...string) {
	for _, m := range methods {
		route.methods = append(route.methods, strings.ToUpper(m))
	}
}

func (route *SimpleRoute) WithContentTypes(contentTypes ...string) {
	for _, c := range contentTypes {
		route.contentTypes = append(route.contentTypes, strings.ToLower(c))
	}
}

func (route *SimpleRoute) Match(r2 RealRoute) (bool, map[string]string) {
	if route.MatchMethod(r2.Method) &&
		route.MatchHost(r2.Host) &&
		route.MatchContentType(r2.ContentType) {
		return route.MatchPath(r2.PathSegments)
	}

	return false, nil
}

// MatchMethod return whether the method is equal to current SimpleRoute
func (route *SimpleRoute) MatchMethod(method string) bool {
	return stringIn(strings.ToUpper(method), route.methods)
}

// MatchHost return whether the host is equal to current SimpleRoute
func (route *SimpleRoute) MatchHost(host string) bool {
	if len(route.hosts) == 0 {
		return true
	}

	return stringIn(strings.ToLower(host), route.hosts)
}

// MatchContentType return whether the Content-Type is equal to current SimpleRoute
func (route *SimpleRoute) MatchContentType(contentType string) bool {
	if len(route.contentTypes) == 0 {
		return true
	}

	return stringIn(strings.ToLower(contentType), route.contentTypes)
}

// MatchPath return whether the path is equal to current SimpleRoute
func (route *SimpleRoute) MatchPath(segments []string) (bool, map[string]string) {
	segmentsLength := len(segments)
	if segmentsLength != (len(route.parsedPaths[ParsedPathPlaceholder]) + len(route.parsedPaths[ParsedPathPlain])) {
		return false, nil
	}

	for _, segment := range route.parsedPaths[ParsedPathPlain] {
		if segmentsLength <= segment.Index {
			return false, nil
		}

		if segments[segment.Index] != segment.Segment {
			return false, nil
		}
	}

	pathVars := make(map[string]string)
	for _, segment := range route.parsedPaths[ParsedPathPlaceholder] {
		if segmentsLength <= segment.Index {
			return false, nil
		}

		pathVars[segment.Segment] = segments[segment.Index]
	}

	return true, pathVars
}

func (route *SimpleRoute) Handle() interface{} {
	return route.handler
}

func (route *SimpleRoute) String() string {
	return fmt.Sprintf(
		"host=%s, method=%s, path=%s, content_type=%s",
		strings.Join(route.hosts, ","),
		strings.Join(route.methods, ","),
		route.path,
		strings.Join(route.contentTypes, ","),
	)
}
