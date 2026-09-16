package mrpc

import (
	"encoding/json"
)

const (
	CodeOK             int32 = 0
	CodeBadRequest     int32 = 100
	CodeMethodNotFound int32 = 101
	CodeError          int32 = 102
)

type Request struct {
	Service string          `json:"service"`
	Method  string          `json:"method"`
	Request json.RawMessage `json:"request"`
}

type Response struct {
	Code     int32           `json:"code"`
	Message  string          `json:"message"`
	Response json.RawMessage `json:"response"`
}

// client
func EncodeRequest(service, method string, req any) ([]byte, error) {

	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	mrpcRequest := Request{
		Service: service,
		Method:  method,
		Request: json.RawMessage(reqData),
	}
	payload, err := json.Marshal(&mrpcRequest)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func DecodeResponse(data []byte) (*Response, error) {
	resp := Response{}
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// server
func DecodeRequest(data []byte) (*Request, error) {

	req := Request{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func EncodeResponse(code int32, message string, resp any) ([]byte, error) {

	respData, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	mrpcResponse := Response{
		Code:     code,
		Message:  message,
		Response: json.RawMessage(respData),
	}
	payload, err := json.Marshal(mrpcResponse)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
