package xprotocol

import (
	"encoding/binary"
	"io"
	. "xthrift"
)

type BinaryProtocol struct {
	r io.Reader
	w io.Writer
}

func NewBinaryProtocol(r io.Reader, w io.Writer) *BinaryProtocol {
	return &BinaryProtocol{
		r: r,
		w: w,
	}
}

func (self *BinaryProtocol) skip(len int) (message []byte) {
	message = make([]byte, len)
	var i, err = io.ReadFull(self.r, message)
	if err == nil {
		self.w.Write(message[:i])
	} else {
		panic(err)
	}
	return
}

func (self *BinaryProtocol) readI8() int8 {
	var message = self.skip(1)
	return int8(message[0])
}

func (self *BinaryProtocol) readI16() int16 {
	var message = self.skip(2)
	return int16(binary.BigEndian.Uint16(message))
}

func (self *BinaryProtocol) readI32() int32 {
	var message = self.skip(4)
	return int32(binary.BigEndian.Uint32(message))
}

func (self *BinaryProtocol) readString() string {
	var size = self.readI32()
	// TODO readLimit
	return string(self.skip(int(size)))
}

func (self *BinaryProtocol) skipString() {
	var size = self.readI32()
	// TODO readLimit
	self.skip(int(size))
}

func (self *BinaryProtocol) skipList() {
	var etype = self.readI8()
	var size = self.readI32()
	for i := int32(0); i < size; i++ {
		self.skipType(etype)
	}
}

func (self *BinaryProtocol) skipMap() {
	var ktype = self.readI8()
	var vtype = self.readI8()
	var size = self.readI32()
	for i := int32(0); i < size; i++ {
		self.skipType(ktype)
		self.skipType(vtype)
	}
}

func (self *BinaryProtocol) skipStruct() {
	for {
		var ftype = self.readI8()
		if ftype == T_STOP {
			return
		}
		// fid
		self.skip(2)
		self.skipType(ftype)
	}
}

func (self *BinaryProtocol) skipType(ttype int8) {
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

func (self *BinaryProtocol) ReadMessageBegin() (
	version int, fname string, seqid int,
) {
	// TODO version is useless.
	version = 0
	self.skip(4)
	fname = self.readString()
	seqid = int(self.readI32())
	return
}

func (self *BinaryProtocol) SkipMessageBody() {
	self.skipStruct()
}
