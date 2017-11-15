package xprotocol

// TODO use BigEndian

import (
	"encoding/binary"
	"io"
	. "xthrift"
	. "xtransport"
)

const (
	VERSION_MASK uint32 = 0xffff0000
	VERSION_1    uint32 = 0x80010000
	TYPE_MASK    uint32 = 0x000000ff
)

const (
	maxMessageNameSize = 128
)

type BinaryProtocol struct {
	itrans      Transport
	otrans      Transport
	direction   bool
	strictRead  bool
	strictWrite bool
	buf         []byte
}

type BinaryProtocolFactory struct {
	strictRead  bool
	strictWrite bool
}

func (self *BinaryProtocolFactory) NewProtocol(t Transport) Protocol {
	return NewTBinaryProtocol(t, self.strictRead, self.strictWrite)
}

func NewTBinaryProtocol(conn Transport, sr, sw bool) *BinaryProtocol {
	return &BinaryProtocol{
		itrans:      conn,
		otrans:      NewTMemoryBuffer(),
		strictRead:  sr,
		strictWrite: sw,
		buf:         make([]byte, 32),
	}
}

func NewTBinaryProtocolFactory(strictRead, strictWrite bool) *BinaryProtocolFactory {
	return &BinaryProtocolFactory{
		strictRead:  strictRead,
		strictWrite: strictWrite,
	}
}

var TBinaryProtocolFactory = NewTBinaryProtocolFactory(true, false)

func (p *BinaryProtocol) Forward() {
	if p.direction {
		p.itrans, p.otrans = p.otrans, p.itrans
		p.direction = false
	}
}

func (p *BinaryProtocol) Reverse() {
	if !p.direction {
		p.itrans, p.otrans = p.otrans, p.itrans
		p.direction = true
	}
}

func (p *BinaryProtocol) SkipStruct() {
	p.skipStruct()
}

func (p *BinaryProtocol) ReadMessageBegin() (
	fname string, mtype byte, seqid int32,
) {
	size := p.readI32()
	if size < 0 {
		version := uint32(size) & VERSION_MASK
		if version != VERSION_1 {
			panic(ProtocolError{"BinaryProtocol", "bad version in ReadMessageBegin"})
		}
		mtype = byte(uint32(size) & TYPE_MASK)
		fname = p.readString()
	} else {
		panic(ProtocolError{"BinaryProtocol", "no protocol version header"})
	}
	seqid = p.readI32()
	return
}

func (p *BinaryProtocol) ReadMessageEnd() {
}

func (p *BinaryProtocol) WriteMessageBegin(fname string, mtype byte, seqid int32) {
	version := int32(VERSION_1 | uint32(mtype))
	p.writeI32(version)
	p.writeString(fname)
	p.writeI32(seqid)
}

func (p *BinaryProtocol) WriteMessageEnd() {
	p.otrans.Flush()
}

func (p *BinaryProtocol) SetServerTransport(trans Transport) {
	p.otrans = trans
}

func (p *BinaryProtocol) GetServerTransport() Transport {
	return p.otrans
}

func (self *BinaryProtocol) skip(len uint) {
	b := self.buf
	if len > 32 {
		b = make([]byte, len)
	} else {
		b = b[:len]
	}

	// what will happened if length is not enough?
	if _, err := io.ReadFull(self.itrans, b); err != nil {
		panic(err)
	} else {
		self.otrans.Write(b)
	}
}

func (self *BinaryProtocol) read(len uint) []byte {
	b := self.buf
	if len > 32 {
		b = make([]byte, len)
	} else {
		b = b[:len]
	}

	// what will happened if length is not enough?
	if _, err := io.ReadFull(self.itrans, b); err != nil {
		panic(err)
	}
	return b
}

func (self *BinaryProtocol) readByte() byte {
	var message = self.read(1)
	return message[0]
}

func (self *BinaryProtocol) readI32() int32 {
	var message = self.read(4)
	return int32(binary.BigEndian.Uint32(message))
}

func (self *BinaryProtocol) readString() string {
	var size = self.readI32()
	return string(self.read(uint(size)))
}

func (self *BinaryProtocol) write(message []byte) {
	if _, err := self.otrans.Write(message); err != nil {
		panic(err)
	}
}

func (self *BinaryProtocol) writeByte(v byte) {
	b := self.buf
	b[0] = v
	self.write(b[:1])
}

func (self *BinaryProtocol) writeI32(v int32) {
	b := self.buf
	binary.BigEndian.PutUint32(b, uint32(v))
	self.write(b[:4])
}

func (self *BinaryProtocol) writeString(v string) {
	len := len(v)
	self.writeI32(int32(len))

	b := self.buf
	if len > 32 {
		b = []byte(v)
		self.write(b)
	} else {
		copy(b, v)
		self.write(b[:len])
	}
}

func (self *BinaryProtocol) skipString() {
	var size = self.readI32()
	self.writeI32(size)
	self.skip(uint(size))
}

func (self *BinaryProtocol) skipList() {
	var etype = self.readByte()
	var size = self.readI32()
	self.writeByte(etype)
	self.writeI32(size)
	for i := int32(0); i < size; i++ {
		self.skipType(etype)
	}
}

func (self *BinaryProtocol) skipMap() {
	var ktype = self.readByte()
	var vtype = self.readByte()
	var size = self.readI32()
	self.writeByte(ktype)
	self.writeByte(vtype)
	self.writeI32(size)
	for i := int32(0); i < size; i++ {
		self.skipType(ktype)
		self.skipType(vtype)
	}
}

func (self *BinaryProtocol) skipStruct() {
	for {
		var ftype = self.readByte()
		self.writeByte(ftype)
		if ftype == T_STOP {
			return
		}
		// fid
		self.skip(2)
		self.skipType(ftype)
	}
}

func (self *BinaryProtocol) skipType(ttype byte) {
	switch ttype {
	case T_BOOL:
		self.skip(1)
	case T_I08:
		self.skip(1)
	case T_I16:
		self.skip(2)
	case T_I32:
		self.skip(4)
	case T_I64:
		self.skip(8)
	case T_DOUBLE:
		self.skip(8)
	case T_STRING:
		self.skipString()
	case T_SET:
		self.skipList()
	case T_LIST:
		self.skipList()
	case T_MAP:
		self.skipMap()
	case T_STRUCT:
		self.skipStruct()
	default:
		panic("unsupported field type")
	}
}
