package common

import (
	"encoding/json"
	"net/http"
	"prolific/config"
	"prolific/debug"
)

type Error struct {
	Code	int 	`json:"code"`
	Reason	string	`json:"reason"`
}

type Response struct {
	Success	bool     `json:"success"`
	Error 	*Error     `json:"error,omitempty"`
	Message	string   `json:"message,omitempty"`
	Data	interface{} `json:"data,omitempty"`
}

func (response *Response) SetError(error *Error) *Response {
	response.Success = false
	response.Error = error
	return response
}

func CreateError(code int, reason string) *Error {
	return &Error{code, reason }
}

func CreateResponse() *Response {
	return &Response{ Success: true }
}

func SendResponse(writer http.ResponseWriter, response *Response) {
	SendResponseWithStatusCode(writer, response, http.StatusOK)
}

func SendResponseWithStatusCode(writer http.ResponseWriter, response *Response, statusCode int) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		debug.Println(err)
	}
}

func NotFoundHandler(writer http.ResponseWriter, request *http.Request) {
	response := CreateResponse()

	handleNotFound := config.GetWithDefault("Server", "Handle_Not_Found", "true") == "true"
	if handleNotFound {
		statusCode := http.StatusNotFound
		response.SetError(CreateError(statusCode, http.StatusText(statusCode)))
		SendResponseWithStatusCode(writer, response, statusCode)
	} else {
		SendResponse(writer, response)
	}
}
