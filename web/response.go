package web

import "net/http"

// simpleResponser is a response object which wrap http.ResponseWriter
type simpleResponser struct {
	w        http.ResponseWriter
	headers  map[string][]string
	cookie   *http.Cookie
	original []byte
	code     int
}

func NewResponseCreator(w http.ResponseWriter) Responsor {
	return &simpleResponser{
		w:       w,
		headers: make(map[string][]string),
	}
}

func (resp *simpleResponser) Raw() http.ResponseWriter {
	return resp.w
}

// GetCode get response code
func (resp *simpleResponser) GetCode() int {
	return resp.code
}

// SetCode set response code
func (resp *simpleResponser) SetCode(code int) {
	resp.code = code
}

// ResponseWriter return the http.ResponseWriter
func (resp *simpleResponser) ResponseWriter() http.ResponseWriter {
	return resp.w
}

// SetContent set response content
func (resp *simpleResponser) SetContent(content []byte) {
	resp.original = content
}

// Header set response header
func (resp *simpleResponser) Header(key string, values ...string) {
	resp.headers[key] = values
}

// Cookie set cookie
func (resp *simpleResponser) Cookie(cookie *http.Cookie) {
	// http.SetCookie(resp.w, cookie)
	resp.cookie = cookie
}

// Flush send all response contents to client
func (resp *simpleResponser) Flush() {
	// set response headers
	for key, value := range resp.headers {
		for _, v := range value {
			resp.w.Header().Add(key, v)
		}
	}

	// set cookies
	if resp.cookie != nil {
		http.SetCookie(resp.w, resp.cookie)
	}

	// set response code
	resp.w.WriteHeader(resp.code)

	// send response body
	_, _ = resp.w.Write(resp.original)
}

// M represents a kv response items
type M map[string]interface{}
