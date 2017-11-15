package xprocessor

import (
	"net"
	"strings"
	"xhandler"
	"xlog"
	. "xprotocol"
	. "xtransport"
)

type MultiplexedProcessor struct {
	pf   ProtocolFactory
	hmap map[string]*xhandler.Handler
}

func (self *MultiplexedProcessor) Add(
	name string,
	handler *xhandler.Handler,
) (err error) {
	self.hmap[name] = handler
	return
}

func (self *MultiplexedProcessor) forward(protocol Protocol) {
	name, mtype, seqid := protocol.ReadMessageBegin()

	segments := strings.SplitN(name, ":", 2)
	if len(segments) != 2 {
		err := NewProcessorError("fname format mismatch: %s", name)
		panic(err)
	}
	service, fname := segments[0], segments[1]
	xlog.Info("service, fname = %s, %s", service, fname)

	handler := self.hmap[service]
	if handler == nil {
		err := NewProcessorError("no handler has been set for: %s", service)
		panic(err)
	}

	otrans, err := handler.GetTransport()
	if err != nil {
		panic(err)
	}
	protocol.SetServerTransport(otrans)
	protocol.WriteMessageBegin(fname, mtype, seqid)
	protocol.SkipStruct()
	protocol.ReadMessageEnd()
	protocol.WriteMessageEnd()
	protocol.Reverse()
}

func (self *MultiplexedProcessor) reverse(protocol Protocol) {
	name, mtype, seqid := protocol.ReadMessageBegin()
	protocol.WriteMessageBegin(name, mtype, seqid)
	protocol.SkipStruct()
	protocol.ReadMessageEnd()
	protocol.WriteMessageEnd()
	protocol.Forward()
}

func (self *MultiplexedProcessor) Process(conn net.Conn) {
	itrans := NewTSocketConn(conn)
	protocol := self.pf.NewProtocol(itrans)

	defer itrans.Close()

	for {
		xlog.Debug("multiplexed forward")
		self.forward(protocol)
		xlog.Debug("multiplexed reverse")
		self.reverse(protocol)
		// TODO Should ServerTransport close at any time?
		protocol.GetServerTransport().Close()
	}

}

func NewMultiplexedProcessor(pf ProtocolFactory) *MultiplexedProcessor {
	return &MultiplexedProcessor{
		pf:   pf,
		hmap: make(map[string]*xhandler.Handler),
	}
}
