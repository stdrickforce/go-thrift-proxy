package xhandler

import (
	"io"
	"net/http"
)

type HttpHandler struct {
	uri string
}

func NewHttpHandler(uri string) HttpHandler {
	return HttpHandler{
		uri: uri,
	}
}

func (self HttpHandler) Handle(r io.Reader) io.ReadCloser {
	resp, err := http.Post(self.uri, "application/thrift", r)
	if err != nil {
		panic(err)
	}
	return resp.Body
}
