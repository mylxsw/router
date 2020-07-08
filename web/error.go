package web


// Error is a interface for http error
type Error interface {
	Error() string
	StatusCode() int
}

// JSONAble identify a value can convert to json object
type JSONAble interface {
	// ToJSON convert a value to jsonable struct or map/array/slice etc
	ToJSON() interface{}
}

// PlainError is a error object which implements Error interface
type PlainError struct {
	err  error
	code int
}

// WrapPlainError warps a error only have message and code
func WrapPlainError(err error, code int) PlainError {
	return PlainError{
		err:  err,
		code: code,
	}
}

func (p PlainError) Error() string {
	return p.err.Error()
}

func (p PlainError) StatusCode() int {
	return p.code
}

// JSONError is a error object which implements Error and JSONAble interface
type JSONError struct {
	err  error
	code int
}

// WrapJSONError wrap a error to JSONError
func WrapJSONError(err error, code int) JSONError {
	return JSONError{err: err, code: code}
}

func (apiErr JSONError) Error() string {
	return apiErr.err.Error()
}

func (apiErr JSONError) StatusCode() int {
	return apiErr.code
}

func (apiErr JSONError) ToJSON() interface{} {
	return M{
		"error": apiErr.err.Error(),
	}
}
