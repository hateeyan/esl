package esl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
)

type Client struct {
	conn   *net.TCPConn
	logger Logger
}

func Dial(addr, password string) (*Client, error) {
	raddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}
	client := &Client{conn: conn, logger: defaultLogger}
	if err := client.authenticate(password); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return client, nil
}

func (c *Client) Run(fn func(msg *Message)) {
	r := bufio.NewReader(c.conn)
	for {
		msg := c.parseMessage(r)
		go func() {
			fn(msg)
			releaseMessage(msg)
		}()
	}
}

func (c *Client) parseMessage(r *bufio.Reader) *Message {
	e := acquireMessage()
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			c.logger.Printf("unable to read header from socket: %v", err)
			break
		}

		i := bytes.IndexByte(line, ':')
		if i == -1 {
			continue
		}
		e.Header.Add(line[:i], bytes.TrimSpace(line[i+1:len(line)-1]))

		peek, _ := r.Peek(1)
		if bytes.Compare(peek, []byte{'\n'}) == 0 {
			_, _ = r.Discard(1)
			n, err := e.Header.GetInt("Content-Length")
			if err != nil {
				c.logger.Printf("unable to parse int value: %v", err)
			}
			// has body
			if n > 0 {
				if len(e.body) < n {
					e.body = make([]byte, n)
				} else {
					e.body = e.body[:n]
				}
				_, err := r.Read(e.body)
				if err != nil && err != io.EOF {
					c.logger.Printf("unable to read body from socket: %v", err)
				}
			}
			break
		}
	}
	return e
}

func (c *Client) authenticate(password string) error {
	r := bufio.NewReader(c.conn)
	msg := c.parseMessage(r)
	ct := msg.ContentType()
	if ct != "auth/request" {
		return fmt.Errorf("unexpected Content-Type for authenticate: %s", ct)
	}
	releaseMessage(msg)

	_, err := c.conn.Write([]byte("auth " + password + "\n\n"))
	if err != nil {
		return err
	}

	msg = c.parseMessage(r)
	ct = msg.ContentType()
	reply := msg.Header.Get("Reply-Text")
	releaseMessage(msg)
	if ct == "command/reply" && reply == "+OK accepted" {
		return nil
	}

	return fmt.Errorf("login failed: %s", reply)
}

func (c *Client) Send(command string) error {
	_, err := c.conn.Write([]byte(command + "\n\n"))
	return err
}

func (c *Client) Close() error {
	return c.conn.Close()
}
