package esl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var ErrAclDenied = errors.New("access denied, please check acl config")

type Handler func(msg *Message)

type Client struct {
	addr     *net.TCPAddr
	password string
	r        *bufio.Reader
	wc       io.WriteCloser

	OnReconnect func(c *Client)

	done             bool
	authChan         chan error
	mu               sync.Mutex
	commandReplyChan chan *CommandReply
}

func Dial(addr, password string, fn Handler) (*Client, error) {
	raddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	client := &Client{addr: raddr, password: password, authChan: make(chan error, 1), commandReplyChan: make(chan *CommandReply, 8)}
	if err = client.dial(); err != nil {
		return nil, err
	}
	go client.run(fn)
	err = client.waitAuth()
	return client, err
}

func (c *Client) run(fn Handler) {
	for {
		c.waitMessage(fn)

		if c.done {
			break
		}
		c.reconnect()
	}
}

func (c *Client) dial() error {
	conn, err := net.DialTCP("tcp", nil, c.addr)
	if err != nil {
		return err
	}
	logger.Printf("connected to fs esl: %s", c.addr.String())
	if c.wc != nil {
		_ = c.wc.Close()
	}
	c.wc = conn
	if c.r == nil {
		c.r = bufio.NewReader(conn)
	} else {
		c.r.Reset(conn)
	}
	return nil
}

func (c *Client) waitMessage(fn Handler) {
	for {
		msg, err := parseMessage(c.r)
		if err != nil {
			logger.Printf("unable to parse message: %v", err)
			break
		}
		switch msg.ContentType() {
		case commandReply:
			reply := <-c.commandReplyChan
			reply.parse(msg)
			releaseMessage(msg)
		case authRequest:
			go c.authenticate(c.password)
		case eventPlain:
			go func() {
				fn(msg.payload())
				releaseMessage(msg)
			}()
		case rudeRejection:
			c.authChan <- ErrAclDenied
			releaseMessage(msg)
			return
		}
	}
}

func parseMessage(r *bufio.Reader) (*Message, error) {
	e := acquireMessage()
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			return nil, err
		}

		i := bytes.IndexByte(line, ':')
		if i == -1 {
			continue
		}
		e.Header.Add(line[:i], line[i+2:len(line)-1])

		peek, _ := r.Peek(1)
		if bytes.Compare(peek, []byte{'\n'}) == 0 {
			_, _ = r.Discard(1)
			n, err := e.Header.ContentLength()
			if err != nil {
				logger.Printf("unable to parse Content-Length: %v", err)
			}
			// has body
			if n > 0 {
				if len(e.body) < n {
					e.body = make([]byte, n)
				}
				_, err = io.ReadFull(r, e.body)
				if err != nil {
					return nil, err
				}
			}
			break
		}
	}
	return e, nil
}

func (c *Client) authenticate(password string) {
	reply := newCommandReply()

	err := c.sendCommand("auth "+password, &reply)
	if err != nil {
		c.authChan <- err
		return
	}
	reply.wait()
	if reply.Succeed() {
		logger.Printf("login to fs esl: %s", c.addr.String())
		c.authChan <- nil
	} else {
		c.authChan <- fmt.Errorf("authenticate (password: %s) failed", c.password)
	}
}

func (c *Client) reconnect() {
	for {
		if err := c.dial(); err != nil {
			logger.Printf("unable to connect to fs esl: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	go func() {
		if err := c.waitAuth(); err != nil {
			logger.Printf("fs esl reconnect failed: %v", err)
			return
		}
		logger.Printf("fs esl reconnected: %s", c.addr.String())
		if c.OnReconnect != nil {
			c.OnReconnect(c)
		}
	}()
}

func (c *Client) waitAuth() error {
	select {
	case err := <-c.authChan:
		return err
	case <-time.After(3 * time.Second):
		return fmt.Errorf("authenticate timeout")
	}
}

func (c *Client) sendCommand(command string, reply *CommandReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.send(command); err != nil {
		return err
	}
	c.commandReplyChan <- reply
	return nil
}

func (c *Client) send(command string) error {
	_, err := c.wc.Write([]byte(command + "\n\n"))
	return err
}

func (c *Client) Command(command string) CommandReply {
	reply := newCommandReply()

	err := c.sendCommand(command, &reply)
	if err != nil {
		reply.err = err
		return reply
	}
	reply.wait()

	return reply
}

func (c *Client) Close() error {
	c.done = true
	return c.wc.Close()
}
