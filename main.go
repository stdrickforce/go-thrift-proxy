package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"xhandler"
	"xlog"
	"xprotocol"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	ini "gopkg.in/ini.v1"
)

type Handler struct {
	name string
	h    xhandler.Handler
}

type Gateway struct {
	registry map[string]Handler
	wg       sync.WaitGroup
}

func (self *Gateway) process(conn net.Conn, h xhandler.Handler) error {
	defer conn.Close()
	defer func() {
		if err := recover(); err != nil {
			if err != io.EOF {
				// TODO report err to sentry.
				fmt.Println(err)
			}
		}
	}()

	var buf = new(bytes.Buffer)
	var p = xprotocol.NewBinaryProtocol(conn, buf)

	for {
		fname, seqid := self.recv_req_header(p)
		xlog.Info(fname)
		_, _ = fname, seqid

		self.recv_req_body(p)

		resp := h.Handle(buf)

		self.send(resp, conn)
	}

}

func (self *Gateway) recv_req_header(
	p *xprotocol.BinaryProtocol,
) (fname string, seqid int) {
	_, fname, seqid = p.ReadMessageBegin()
	return
}

func (self *Gateway) recv_req_body(
	p *xprotocol.BinaryProtocol,
) {
	p.SkipMessageBody()
}

func (self *Gateway) send(r io.Reader, w io.Writer) {
	var buf bytes.Buffer
	buf.ReadFrom(r)
	w.Write(buf.Bytes())
}

func (self *Gateway) Register(name, addr string, h xhandler.Handler) {
	self.registry[addr] = Handler{
		name: name,
		h:    h,
	}
}

func (self *Gateway) Serve() {
	for addr, handler := range self.registry {
		self.wg.Add(1)
		go self.run(addr, handler)
	}
	self.wg.Wait()
}

func (self *Gateway) run(addr string, handler Handler) error {
	// TODO goroutine unexpected exit.
	defer self.wg.Done()

	xlog.Info("%s gateway is listening on %s...", handler.name, addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		// accept connection on port
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go self.process(conn, handler.h)
	}
}

func (self *Gateway) LoadConfig(filepath string) (err error) {
	cfg, err := ini.Load(filepath)
	if err != nil {
		return
	}

	for _, section := range cfg.ChildSections("service") {
		type_ := section.Key("type").String()
		switch type_ {
		case "http":
			self.Register(
				section.Name()[8:],
				section.Key("addr").String(),
				xhandler.NewHttpHandler(
					section.Key("uri").String(),
				),
			)
		default:
			return errors.New(
				fmt.Sprintf("handler type mismatch: %s", type_),
			)
		}
	}
	return
}

func MakeGateway() *Gateway {
	return &Gateway{
		registry: make(map[string]Handler),
	}
}

var (
	config = kingpin.Flag("config", "Config file.").Short('c').Required().String()
)

func main() {
	kingpin.Parse()
	var gateway = MakeGateway()
	if err := gateway.LoadConfig(*config); err != nil {
		panic(err)
	}
	gateway.Serve()
}
