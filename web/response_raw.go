package web

// RawResponse 原生响应
type RawResponse struct {
	response Responsor
}

func (resp *RawResponse) Code() int {
	return resp.response.GetCode()
}

// NewRawResponse create a RawResponse
func NewRawResponse(response Responsor) *RawResponse {
	return &RawResponse{response: response}
}

// response get real response object
func (resp *RawResponse) Response() Responsor {
	return resp.response
}

// Send flush response to client
func (resp *RawResponse) Send() error {
	resp.response.Flush()
	return nil
}
