package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

type ErrorCode int

const (
	ECSuccess ErrorCode = 0
	ECUnknown           = iota + 1000
	ECInvalidParam
	ECJobNotFound
	ECJobNotRunning
)

type JobStatus string

const (
	JSRunning  JobStatus = "running"
	JSCanceled           = "canceled"
	JSFinished           = "finished"
	JSFailed             = "failed"
)

const (
	ContentType     = "Content-Type"
	JsonContentType = "application/json;charset=UTF-8"
)

type Response struct {
	Errno ErrorCode   `json:"errno"`
	Error string      `json:"error"`
	Data  interface{} `json:"data,omitempty"`
}

func NewResponse() *Response {
	return &Response{Errno: ECSuccess, Error: "succeed"}
}

func (o *Response) SetData(data interface{}) *Response {
	o.Data = data
	return o
}

func (o *Response) SetError(errno ErrorCode, error string) *Response {
	o.Errno = errno
	o.Error = error
	return o
}

func ServeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set(ContentType, JsonContentType)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Errorf("Error occured when marshalling response: %s", err)
	}
}
