package esl

import (
	"errors"
	"net"
)

const (
	defaultLocalAddr = ":9090"
)

type OutboundHandler func(conn *Connection)

type Outbound struct {
	// LocalAddr the bind address for outbound socket
	// default: ":9090"
	LocalAddr string
	// Handler handle the new Outbound connection
	Handler OutboundHandler
}

func (o *Outbound) Serve() error {
	if o.Handler == nil {
		return errors.New("unset outbound handler")
	}

	addr := o.LocalAddr
	if addr == "" {
		addr = defaultLocalAddr
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Printf("unable to accept new connection: %v", err)
			continue
		}
		go o.handleOne(conn)
	}
}

func (o *Outbound) handleOne(conn net.Conn) {
	defer conn.Close()
	c := acquireConnection(conn, outbound)
	if err := c.connect(); err != nil {
		logger.Printf("unable to connect to freeswitch: %v", err)
		releaseOutbound(c)
		return
	}
	go o.Handler(c)
	c.waitMessage()
	releaseOutbound(c)
}
