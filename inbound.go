package esl

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	ErrAclDenied  = errors.New("access denied, please check acl config")
	ErrMaxRetried = errors.New("a series of reconnecting have failed")
)

type Handler func(msg *Message)

type Inbound struct {
	// Address freeswitch esl address
	// Required
	// e.g. 192.168.40.249:8021
	Address string
	// Password freeswitch esl auth password
	// Required
	Password string
	// Maximum duration for event socket connected
	DialTimeout time.Duration
	// MaxReconnect max reconnect count
	// If the value is set to 0, then no limit will be enforced
	// Default: 0
	MaxReconnect int
	// Apps event handlers
	// See Applications for more information
	Apps Applications

	// internal
	*Connection
	closed bool
	mu     sync.Mutex
}

func (i *Inbound) Run() error {
	if err := i.dial(i.Password); err != nil {
		return err
	}
	go i.run()
	return nil
}

func (i *Inbound) run() {
	for {
		i.Connection.waitMessage()
		_ = i.Connection.Close()

		if i.closed {
			break
		}

		if err := i.reconnect(); err != nil {
			logger.Printf("esl reconnected failed: %v", err)
			break
		}
	}
}

func (i *Inbound) dial(password string) error {
	conn, err := net.DialTimeout("tcp", i.Address, i.DialTimeout)
	if err != nil {
		return err
	}
	if i.Connection == nil {
		i.Connection = acquireConnection(conn, inbound)
		i.Connection.apps = &i.Apps
	} else {
		i.Connection.reset(conn)
	}

	msg, err := parseMessage(i.Connection.r)
	if err != nil {
		_ = conn.Close()
		return err
	}
	if t := msg.ContentType(); t != authRequest {
		_ = conn.Close()
		return fmt.Errorf("invalid auth request: %s", t)
	}
	ReleaseMessage(msg)
	if err = i.authenticate(password); err != nil {
		_ = conn.Close()
		return err
	}

	logger.Printf("connected to fs esl: %s", i.Address)
	return nil
}

// authenticate auth for inbound mode
func (i *Inbound) authenticate(password string) error {
	_, err := fmt.Fprintf(i.conn, "auth %s\n\n", password)
	if err != nil {
		return fmt.Errorf("unable to send auth request: %v", err)
	}

	msg, err := parseMessage(i.Connection.r)
	if err != nil {
		return err
	}
	reply := msg.Header.Get("Reply-Text")
	ReleaseMessage(msg)
	if !strings.HasPrefix(reply, replyOK) {
		return fmt.Errorf("authenticate (password: %s) failed: %s", password, reply)
	}
	return nil
}

func (i *Inbound) reconnect() error {
	for count := 1; i.MaxReconnect == 0 || count <= i.MaxReconnect; count++ {
		err := i.dial(i.Password)
		if i.apps.OnReconnect != nil {
			i.apps.OnReconnect(i, err)
		}
		if err != nil {
			continue
		}
		return nil
	}
	return ErrMaxRetried
}

func (i *Inbound) Close() error {
	i.closed = true
	return i.Connection.Close()
}
