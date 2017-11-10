package xhandler

import (
	"io"
	"xlog"
)

type Handler interface {
	Handle(r io.Reader) io.ReadCloser
}

func init() {
	xlog.Info("xhandler has been initialized.")
}
