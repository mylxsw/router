package web

import (
	"encoding/json"
	"fmt"
)

// JSONResponse json响应
type JSONResponse struct {
	response Responsor
	original interface{}
	code     int
}

func (resp *JSONResponse) Code() int {
	return resp.code
}

// NewJSONResponse 创建JSONResponse对象
func NewJSONResponse(response Responsor, code int, res interface{}) *JSONResponse {
	return &JSONResponse{
		response: response,
		original: res,
		code:     code,
	}
}

// WithCode set response code and return itself
func (resp *JSONResponse) WithCode(code int) *JSONResponse {
	resp.code = code
	return resp
}

// Send create response
func (resp *JSONResponse) Send() error {
	switch resp.original.(type) {
	case []byte:
		resp.response.SetContent(resp.original.([]byte))
	case string:
		resp.response.SetContent([]byte(resp.original.(string)))
	default:
		res, err := json.Marshal(resp.original)
		if err != nil {
			err = fmt.Errorf("json encode failed: %v [%v]", err, resp.original)

			return err
		}
		resp.response.SetContent(res)
	}

	resp.response.SetCode(resp.code)
	resp.response.Header("Content-Type", "application/json; charset=utf-8")

	resp.response.Flush()
	return nil
}
