package xhandler

import (
	"bytes"
)

type SocketHandler struct {
	host string
	port int
}

func NewSocketHandler(host string, port int) (h *SocketHandler) {
	h = &SocketHandler{
		host: host,
		port: port,
	}
	return
}

func (self *SocketHandler) Send(ch chan []byte, b *bytes.Buffer) {
	// TODO nothing.
}
