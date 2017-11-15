package xprotocol

import (
	"fmt"
	. "xtransport"
)

type ProtocolError struct {
	Protocol string
	Message  string
}

func (e ProtocolError) Error() string {
	return fmt.Sprintf("thrift: [%s] %s", e.Protocol, e.Message)
}

type ProtocolFactory interface {
	NewProtocol(Transport) Protocol
}

type Protocol interface {
	Forward()
	Reverse()

	ReadMessageBegin() (name string, mtype byte, seqid int32)
	ReadMessageEnd()

	WriteMessageBegin(name string, mtype byte, seqid int32)
	WriteMessageEnd()

	SkipStruct()

	SetServerTransport(Transport)
	GetServerTransport() Transport
}
