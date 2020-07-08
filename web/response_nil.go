package web

// NilResponse 空响应
type NilResponse struct {
	response Responsor
}

func (resp *NilResponse) Code() int {
	return resp.response.GetCode()
}

// NewNilResponse create a RawResponse
func NewNilResponse(response Responsor) *NilResponse {
	return &NilResponse{response: response}
}

// response get real response object
func (resp *NilResponse) Response() Responsor {
	return resp.response
}

// Send flush response to client
func (resp *NilResponse) Send() error {
	return nil
}
