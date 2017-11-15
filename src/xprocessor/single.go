package xprocessor

import (
	"errors"
	"net"
	"xhandler"
	"xlog"
	. "xprotocol"
	. "xtransport"
)

type SingleProcessor struct {
	name    string
	handler *xhandler.Handler
	pf      ProtocolFactory
}

func (self *SingleProcessor) Add(
	name string,
	handler *xhandler.Handler,
) (err error) {
	self.name = name
	self.handler = handler
	return
}

func (self *SingleProcessor) forward(protocol Protocol) {
	protocol.Forward()
	name, mtype, seqid := protocol.ReadMessageBegin()
	protocol.WriteMessageBegin(name, mtype, seqid)
	protocol.SkipStruct()
	protocol.ReadMessageEnd()
	protocol.WriteMessageEnd()
}

func (self *SingleProcessor) reverse(protocol Protocol) {
	protocol.Reverse()
	xlog.Debug("reverse")
	name, mtype, seqid := protocol.ReadMessageBegin()
	xlog.Debug("read message begin: %s %d %d", name, mtype, seqid)
	protocol.WriteMessageBegin(name, mtype, seqid)
	xlog.Debug("write message begin")
	protocol.SkipStruct()
	xlog.Debug("skip struct")
	protocol.ReadMessageEnd()
	xlog.Debug("read message end")
	protocol.WriteMessageEnd()
	xlog.Debug("write message end")
}

func (self *SingleProcessor) Process(conn net.Conn) {
	itrans := NewTSocketConn(conn)
	protocol := self.pf.NewProtocol(itrans)

	// Get a server transport
	if self.handler == nil {
		panic(errors.New("no handler has been set"))
	}
	otrans, err := self.handler.GetTransport()
	if err != nil {
		panic(err)
	}

	// close transports after process finished.
	defer itrans.Close()
	defer otrans.Close()

	protocol.SetServerTransport(otrans)
	for {
		self.forward(protocol)
		self.reverse(protocol)
	}
}

func NewProcessor(pf ProtocolFactory) *SingleProcessor {
	return &SingleProcessor{
		pf: pf,
	}
}
