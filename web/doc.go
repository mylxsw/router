package web

import (
	"context"
	"net/http"

	"github.com/mylxsw/container"
)

// Log is a interface for logger
type Log interface {
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type Route interface {
	Match(r2 RealRoute) (bool, map[string]string)

	Handle() interface{}
	String() string
	Methods() []string
	Hosts() []string
	Path() string
	ContentTypes() []string
	Decorators() []HandlerDecorator

	WithHost(hosts ...string)
	WithPath(path string)
	WithMethod(methods ...string)
	WithContentTypes(contentTypes ...string)
	WithDecorators(decors ...HandlerDecorator)
	WithHandler(handler interface{})
	PrependDecorators(decors ...HandlerDecorator)
}

type Context interface {
	JSON(res interface{}) *JSONResponse
	JSONWithCode(res interface{}, code int) *JSONResponse
	API(businessCode string, message string, data interface{}) *JSONResponse
	JSONError(res string, code int) *JSONResponse

	YAML(res interface{}) *YAMLResponse
	Nil() *NilResponse
	Plain() *RawResponse

	HTML(res string) *HTMLResponse
	HTMLWithCode(res string, code int) *HTMLResponse

	Error(res string, code int) *ErrorResponse
	Redirect(location string, code int) *RedirectResponse

	Decode(v interface{}) error
	Unmarshal(v interface{}) error
	UnmarshalYAML(v interface{}) error
	PathVar(key string) string
	PathVars() map[string]string
	Input(key string) string
	JSONGet(keys ...string) string
	InputWithDefault(key string, defaultVal string) string
	ToInt(val string, defaultVal int) int
	ToInt64(val string, defaultVal int64) int64
	ToFloat32(val string, defaultVal float32) float32
	ToFloat64(val string, defaultVal float64) float64
	IntInput(key string, defaultVal int) int
	Int64Input(key string, defaultVal int64) int64
	Float32Input(key string, defaultVal float32) float32
	Float64Input(key string, defaultVal float64) float64
	IsXMLHTTPRequest() bool
	AJAX() bool
	IsJSON() bool
	ContentType() string
	AllHeaders() http.Header
	Headers(key string) []string
	Header(key string) string
	Is(method string) bool
	IsGet() bool
	IsPost() bool
	IsHead() bool
	IsDelete() bool
	IsPut() bool
	IsPatch() bool
	IsOptions() bool
	Method() string
	Body() []byte
	File(key string) (*UploadedFile, error)
	Set(key string, value interface{})
	Get(key string) interface{}
	Context() context.Context
	Cookie(name string) (*http.Cookie, error)
	Cookies() []*http.Cookie
	Request() Request
	Response() Responsor
	Container() container.Container
	Validate(validator Validator, jsonResponse bool)
}

type Request interface {
	Raw() *http.Request
	Decode(v interface{}) error
	Unmarshal(v interface{}) error
	UnmarshalYAML(v interface{}) error
	PathVar(key string) string
	PathVars() map[string]string
	Input(key string) string
	JSONGet(keys ...string) string
	InputWithDefault(key string, defaultVal string) string
	ToInt(val string, defaultVal int) int
	ToInt64(val string, defaultVal int64) int64
	ToFloat32(val string, defaultVal float32) float32
	ToFloat64(val string, defaultVal float64) float64
	IntInput(key string, defaultVal int) int
	Int64Input(key string, defaultVal int64) int64
	Float32Input(key string, defaultVal float32) float32
	Float64Input(key string, defaultVal float64) float64
	File(key string) (*UploadedFile, error)
	IsXMLHTTPRequest() bool
	AJAX() bool
	IsJSON() bool
	ContentType() string
	AllHeaders() http.Header
	Headers(key string) []string
	Header(key string) string
	Is(method string) bool
	IsGet() bool
	IsPost() bool
	IsHead() bool
	IsDelete() bool
	IsPut() bool
	IsPatch() bool
	IsOptions() bool
	Method() string
	Body() []byte
	Set(key string, value interface{})
	Get(key string) interface{}

	Context() context.Context
	Cookie(name string) (*http.Cookie, error)
	Cookies() []*http.Cookie

	Validate(validator Validator, jsonResponse bool)
}

// Validator is an interface for validator
type Validator interface {
	Validate(request Request) error
}

// Response is the response interface
type Response interface {
	Send() error
	Code() int
}

// Controller is a interface for controller
type Controller interface {
	// Register register routes for a controller
	Register(router *Router)
}

// Responsor is a response creator
type Responsor interface {
	Raw() http.ResponseWriter
	SetCode(code int)
	ResponseWriter() http.ResponseWriter
	SetContent(content []byte)
	Header(key string, values ...string)
	Cookie(cookie *http.Cookie)
	GetCode() int
	Flush()
}

type Decoder interface {
	Decode(dst interface{}, src map[string][]string) error
}

type structDecoder struct{}

func (dec structDecoder) Decode(dst interface{}, src map[string][]string) error {
	return nil
}
