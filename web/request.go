package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/buger/jsonparser"
	"github.com/mylxsw/container"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// request 请求对象封装
type httpRequest struct {
	r          *http.Request
	body       []byte
	cc         container.Container
	conf       *Config
	stores     map[string]interface{}
	bodyLoader sync.Once
	pathVars   map[string]string
}

// NewRequest create new Request
func NewRequest(cc container.Container, conf *Config, r *http.Request, pathVars map[string]string) Request {
	if pathVars == nil {
		pathVars = make(map[string]string)
	}

	return &httpRequest{
		r:        r,
		body:     nil,
		cc:       cc,
		conf:     conf,
		stores:   make(map[string]interface{}),
		pathVars: pathVars,
	}
}

// Context returns the request's context
func (req *httpRequest) Context() context.Context {
	return req.r.Context()
}

// Cookie returns the named cookie provided in the request or ErrNoCookie if not found.
// If multiple cookies match the given name, only one cookie will be returned.
func (req *httpRequest) Cookie(name string) (*http.Cookie, error) {
	return req.r.Cookie(name)
}

// Cookies parses and returns the HTTP cookies sent with the request.
func (req *httpRequest) Cookies() []*http.Cookie {
	return req.r.Cookies()
}

// Raw get the underlying http.request
func (req *httpRequest) Raw() *http.Request {
	return req.r
}

// Decode decodes form request to a struct
func (req *httpRequest) Decode(v interface{}) error {
	return req.cc.ResolveWithError(func(decoder Decoder) error {
		if req.ContentType() == "multipart/form-data" {
			if err := req.r.ParseMultipartForm(req.conf.MultipartFormMaxMemory); err != nil {
				return errors.Wrap(err, "parse multipart form failed")
			}

			if err := decoder.Decode(v, req.r.MultipartForm.Value); err != nil {
				return errors.Wrap(err, "decode multipart-form failed")
			}

			return nil
		}

		if err := req.r.ParseForm(); err != nil {
			return errors.Wrap(err, "parse form failed")
		}

		if err := decoder.Decode(v, req.r.Form); err != nil {
			return errors.Wrap(err, "decode form failed")
		}

		return nil
	})
}

// Unmarshal unmarshal request body as json object
// result must be reference to a variable
func (req *httpRequest) Unmarshal(v interface{}) error {
	return json.Unmarshal(req.Body(), v)
}

// UnmarshalYAML unmarshal request body as yaml object
// result must be reference to a variable
func (req *httpRequest) UnmarshalYAML(v interface{}) error {
	return yaml.Unmarshal(req.Body(), v)
}

// Set 设置一个变量，存储到当前请求
func (req *httpRequest) Set(key string, value interface{}) {
	req.stores[key] = value
}

// Get 从当前请求提取设置的变量
func (req *httpRequest) Get(key string) interface{} {
	return req.stores[key]
}

// Clear clear all variables in request
func (req *httpRequest) Clear() {
	req.stores = make(map[string]interface{})
}

// HTTPRequest return a http.request
func (req *httpRequest) HTTPRequest() *http.Request {
	return req.r
}

// PathVar return a path parameter
func (req *httpRequest) PathVar(key string) string {
	return req.pathVars[key]
}

// PathVars return all path parameters
func (req *httpRequest) PathVars() map[string]string {
	return req.pathVars
}

// Input return form parameter from request
func (req *httpRequest) Input(key string) string {
	if req.IsJSON() {
		val := req.JSONGet(key)
		if val != "" {
			return val
		}
	}

	return req.r.FormValue(key)
}

func (req *httpRequest) JSONGet(keys ...string) string {
	value, dataType, _, err := jsonparser.Get(req.Body(), keys...)
	if err != nil {
		return ""
	}

	switch dataType {
	case jsonparser.String:
		if res, err := jsonparser.ParseString(value); err == nil {
			return res
		}
	case jsonparser.Number:
		if res, err := jsonparser.ParseFloat(value); err == nil {
			return strconv.FormatFloat(res, 'f', -1, 32)
		}
		if res, err := jsonparser.ParseInt(value); err == nil {
			return fmt.Sprintf("%d", res)
		}
	case jsonparser.Object:
		fallthrough
	case jsonparser.Array:
		return fmt.Sprintf("%x", value)
	case jsonparser.Boolean:
		if res, err := jsonparser.ParseBoolean(value); err == nil {
			if res {
				return "true"
			} else {
				return "false"
			}
		}
	case jsonparser.NotExist:
		fallthrough
	case jsonparser.Null:
		fallthrough
	case jsonparser.Unknown:
		return ""
	}

	return ""
}

// InputWithDefault return a form parameter with a default value
func (req *httpRequest) InputWithDefault(key string, defaultVal string) string {
	val := req.Input(key)
	if val == "" {
		return defaultVal
	}

	return val
}

func (req *httpRequest) ToInt(val string, defaultVal int) int {
	res, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}

	return res
}

func (req *httpRequest) ToInt64(val string, defaultVal int64) int64 {
	res, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultVal
	}

	return res
}

func (req *httpRequest) ToFloat32(val string, defaultVal float32) float32 {
	res, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return defaultVal
	}

	return float32(res)
}

func (req *httpRequest) ToFloat64(val string, defaultVal float64) float64 {
	res, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}

	return res
}

// IntInput return a integer form parameter
func (req *httpRequest) IntInput(key string, defaultVal int) int {
	return req.ToInt(req.Input(key), defaultVal)
}

// Int64Input return a integer form parameter
func (req *httpRequest) Int64Input(key string, defaultVal int64) int64 {
	return req.ToInt64(req.Input(key), defaultVal)
}

// Float32Input return a float32 form parameter
func (req *httpRequest) Float32Input(key string, defaultVal float32) float32 {
	return req.ToFloat32(req.Input(key), defaultVal)
}

// Float64Input return a float64 form parameter
func (req *httpRequest) Float64Input(key string, defaultVal float64) float64 {
	return req.ToFloat64(req.Input(key), defaultVal)
}

// File Retrieving Uploaded Files
func (req *httpRequest) File(key string) (*UploadedFile, error) {
	file, header, err := req.r.FormFile(key)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	tempFile, err := ioutil.TempFile(req.conf.TempDir, req.conf.TempFilePattern)
	if err != nil {
		return nil, fmt.Errorf("can not create temporary file %s", err.Error())
	}
	defer func() {
		_ = tempFile.Close()
	}()

	if _, err := io.Copy(tempFile, file); err != nil {
		return nil, err
	}

	return &UploadedFile{
		Header:   header,
		SavePath: tempFile.Name(),
	}, nil
}

// IsXMLHTTPRequest return whether the request is a ajax request
func (req *httpRequest) IsXMLHTTPRequest() bool {
	return req.r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// AJAX return whether the request is a ajax request
func (req *httpRequest) AJAX() bool {
	return req.IsXMLHTTPRequest()
}

// IsJSON return whether the request is a json request
func (req *httpRequest) IsJSON() bool {
	return req.ContentType() == "application/json"
}

// ContentType return content type for request
func (req *httpRequest) ContentType() string {
	t := req.r.Header.Get("Content-Type")
	if t == "" {
		return "text/html"
	}

	return strings.ToLower(strings.Split(t, ";")[0])
}

// AllHeaders return all http request headers
func (req *httpRequest) AllHeaders() http.Header {
	return req.r.Header
}

// Headers gets all values associated with given key
func (req *httpRequest) Headers(key string) []string {
	res, ok := req.r.Header[key]
	if !ok {
		return make([]string, 0)
	}

	return res
}

// Header gets the first value associated with the given key.
func (req *httpRequest) Header(key string) string {
	return req.r.Header.Get(key)
}

// Is 判断请求方法
func (req *httpRequest) Is(method string) bool {
	return req.Method() == method
}

// IsGet 判断是否是Get请求
func (req *httpRequest) IsGet() bool {
	return req.Is("GET")
}

// IsPost 判断是否是Post请求
func (req *httpRequest) IsPost() bool {
	return req.Is("POST")
}

// IsHead 判断是否是HEAD请求
func (req *httpRequest) IsHead() bool {
	return req.Is("HEAD")
}

// IsDelete 判断是是否是Delete请求
func (req *httpRequest) IsDelete() bool {
	return req.Is("DELETE")
}

// IsPut 判断是否是Put请求
func (req *httpRequest) IsPut() bool {
	return req.Is("PUT")
}

// IsPatch 判断是否是Patch请求
func (req *httpRequest) IsPatch() bool {
	return req.Is("PATCH")
}

// IsOptions 判断是否是Options请求
func (req *httpRequest) IsOptions() bool {
	return req.Is("OPTIONS")
}

// Method 获取请求方法
func (req *httpRequest) Method() string {
	return req.r.Method
}

// Body return request body
func (req *httpRequest) Body() []byte {
	req.bodyLoader.Do(func() {
		body, _ := ioutil.ReadAll(req.r.Body)
		_ = req.r.Body.Close()
		req.r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		req.body = body
	})

	return req.body
}

// Validate execute a validator, if there has an error, panic error to framework
func (req *httpRequest) Validate(validator Validator, jsonResponse bool) {
	if err := validator.Validate(req); err != nil {
		if jsonResponse {
			panic(WrapJSONError(fmt.Errorf("invalid request: %v", err), http.StatusUnprocessableEntity))
		} else {
			panic(WrapPlainError(fmt.Errorf("invalid request: %v", err), http.StatusUnprocessableEntity))
		}
	}
}

// UploadedFile 上传的文件
type UploadedFile struct {
	Header   *multipart.FileHeader
	SavePath string
}

// Extension get the file's extension.
func (file *UploadedFile) Extension() string {
	segs := strings.Split(file.Header.Filename, ".")
	return segs[len(segs)-1]
}

// Store store the uploaded file on a filesystem disk.
func (file *UploadedFile) Store(path string) error {
	if err := os.Rename(file.SavePath, path); err != nil {
		return err
	}

	file.SavePath = path
	return nil
}

// Delete 删除文件
func (file *UploadedFile) Delete() error {
	return os.Remove(file.SavePath)
}

// Name 获取上传的文件名
func (file *UploadedFile) Name() string {
	return file.Header.Filename
}

// Size 获取文件大小
func (file *UploadedFile) Size() int64 {
	return file.Header.Size
}

// GetTempFilename 获取文件临时保存的地址
func (file *UploadedFile) GetTempFilename() string {
	return file.SavePath
}
