package xhandler

import (
	"io"
)

type Handler interface {
	Handle(r io.Reader) io.ReadCloser
}

func init() {
}
