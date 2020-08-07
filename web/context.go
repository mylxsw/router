package web

import (
	"context"
	"net/http"

	"github.com/mylxsw/container"
)

type webContext struct {
	responsor Responsor
	request   Request
	cc        container.Container
	conf      *Config
}

func NewWebContext(router *Router, pathVars map[string]string, writer http.ResponseWriter, request *http.Request) Context {
	return &webContext{
		cc:        router.cc,
		responsor: NewResponseCreator(writer),
		request:   NewRequest(router.cc, router.conf, request, pathVars),
	}
}

func (w *webContext) JSON(res interface{}) *JSONResponse {
	return NewJSONResponse(w.responsor, http.StatusOK, res)
}

func (w *webContext) YAML(res interface{}) *YAMLResponse {
	return NewYAMLResponse(w.responsor, http.StatusOK, res)
}

func (w *webContext) JSONWithCode(res interface{}, code int) *JSONResponse {
	return NewJSONResponse(w.responsor, code, res)
}

func (w *webContext) Nil() *NilResponse {
	return NewNilResponse(w.responsor)
}

func (w *webContext) API(businessCode string, message string, data interface{}) *JSONResponse {
	return w.JSON(struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{
		Code:    businessCode,
		Message: message,
		Data:    data,
	})
}

func (w *webContext) Plain() *RawResponse {
	return NewRawResponse(w.responsor)
}

func (w *webContext) HTML(res string) *HTMLResponse {
	return NewHTMLResponse(w.responsor, http.StatusOK, res)
}

func (w *webContext) HTMLWithCode(res string, code int) *HTMLResponse {
	return NewHTMLResponse(w.responsor, code, res)
}

func (w *webContext) Error(res string, code int) *ErrorResponse {
	return NewErrorResponse(w.responsor, res, code)
}

func (w *webContext) JSONError(res string, code int) *JSONResponse {
	return w.JSONWithCode(M{"error": res}, code)
}

func (w *webContext) Redirect(location string, code int) *RedirectResponse {
	return NewRedirectResponse(w.responsor, w.request, location, code)
}

func (w *webContext) Decode(v interface{}) error {
	return w.request.Decode(v)
}

func (w *webContext) Unmarshal(v interface{}) error {
	return w.request.Unmarshal(v)
}

func (w *webContext) UnmarshalYAML(v interface{}) error {
	return w.request.UnmarshalYAML(v)
}

func (w *webContext) PathVar(key string) string {
	return w.request.PathVar(key)
}

func (w *webContext) PathVars() map[string]string {
	return w.request.PathVars()
}

func (w *webContext) Input(key string) string {
	return w.request.Input(key)
}

func (w *webContext) JSONGet(keys ...string) string {
	return w.request.JSONGet(keys...)
}

func (w *webContext) InputWithDefault(key string, defaultVal string) string {
	return w.request.InputWithDefault(key, defaultVal)
}

func (w *webContext) ToInt(val string, defaultVal int) int {
	return w.request.ToInt(val, defaultVal)
}

func (w *webContext) ToInt64(val string, defaultVal int64) int64 {
	return w.request.ToInt64(val, defaultVal)
}

func (w *webContext) ToFloat32(val string, defaultVal float32) float32 {
	return w.request.ToFloat32(val, defaultVal)
}

func (w *webContext) ToFloat64(val string, defaultVal float64) float64 {
	return w.request.ToFloat64(val, defaultVal)
}

func (w *webContext) IntInput(key string, defaultVal int) int {
	return w.request.IntInput(key, defaultVal)
}

func (w *webContext) Int64Input(key string, defaultVal int64) int64 {
	return w.request.Int64Input(key, defaultVal)
}

func (w *webContext) Float32Input(key string, defaultVal float32) float32 {
	return w.request.Float32Input(key, defaultVal)
}

func (w *webContext) Float64Input(key string, defaultVal float64) float64 {
	return w.request.Float64Input(key, defaultVal)
}

func (w *webContext) IsXMLHTTPRequest() bool {
	return w.request.IsXMLHTTPRequest()
}

func (w *webContext) AJAX() bool {
	return w.request.AJAX()
}

func (w *webContext) IsJSON() bool {
	return w.request.IsJSON()
}

func (w *webContext) ContentType() string {
	return w.request.ContentType()
}

func (w *webContext) AllHeaders() http.Header {
	return w.request.AllHeaders()
}

func (w *webContext) Headers(key string) []string {
	return w.request.Headers(key)
}

func (w *webContext) Header(key string) string {
	return w.request.Header(key)
}

func (w *webContext) Is(method string) bool {
	return w.request.Is(method)
}

func (w *webContext) IsGet() bool {
	return w.request.IsGet()
}

func (w *webContext) IsPost() bool {
	return w.request.IsPost()
}

func (w *webContext) IsHead() bool {
	return w.request.IsHead()
}

func (w *webContext) IsDelete() bool {
	return w.request.IsDelete()
}

func (w *webContext) IsPut() bool {
	return w.request.IsPut()
}

func (w *webContext) IsPatch() bool {
	return w.request.IsPatch()
}

func (w *webContext) IsOptions() bool {
	return w.request.IsOptions()
}

func (w *webContext) Method() string {
	return w.request.Method()
}

func (w *webContext) Body() []byte {
	return w.request.Body()
}

func (w *webContext) File(key string) (*UploadedFile, error) {
	return w.request.File(key)
}

func (w *webContext) Set(key string, value interface{}) {
	w.request.Set(key, value)
}

func (w *webContext) Get(key string) interface{} {
	return w.request.Get(key)
}

func (w *webContext) Context() context.Context {
	return w.request.Context()
}

func (w *webContext) Cookie(name string) (*http.Cookie, error) {
	return w.request.Cookie(name)
}

func (w *webContext) Cookies() []*http.Cookie {
	return w.request.Cookies()
}

func (w *webContext) Request() Request {
	return w.request
}

func (w *webContext) Response() Responsor {
	return w.responsor
}

func (w *webContext) Container() container.Container {
	return w.cc
}

func (w *webContext) Validate(validator Validator, jsonResponse bool) {
	w.request.Validate(validator, jsonResponse)
}
