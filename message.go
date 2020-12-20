package esl

import (
	"bytes"
	"strings"
	"sync"
)

var messagePool sync.Pool

type Message struct {
	Header Header

	// bs body start
	bs   int
	body []byte
}

func NewMessage() *Message {
	return &Message{Header: Header{contentLength: -1}}
}

func acquireMessage() *Message {
	got := messagePool.Get()
	if got == nil {
		return NewMessage()
	}
	msg := got.(*Message)
	msg.reset()
	return msg
}

func releaseMessage(e *Message) {
	messagePool.Put(e)
}

func (e *Message) Body() []byte {
	n, _ := e.Header.ContentLength()
	return e.body[e.bs:n]
}

func (e *Message) ContentType() string {
	return e.Header.Get("Content-Type")
}

func (e *Message) payload() *Message {
	buf := e.Body()
	e.Header.kvs = e.Header.kvs[:0]
	var bs int
	for {
		l := bytes.IndexByte(buf, '\n')
		if l == -1 {
			break
		}
		i := bytes.IndexByte(buf, ':')
		if i != -1 {
			e.Header.Add(buf[:i], buf[i+2:l])
		}

		if len(buf) >= l+1 {
			if buf[l+1] != '\n' {
				buf = buf[l+1:]
				bs += l + 1
				continue
			}

			e.bs = bs + l + 2
			break
		}
	}
	return e
}

func (e *Message) reset() {
	e.bs = 0
	e.Header.reset()
}

type CommandReply struct {
	replyText string
	jobUUID   string
	err       error

	c chan struct{}
}

func newCommandReply() CommandReply {
	return CommandReply{c: make(chan struct{})}
}

func (c *CommandReply) parse(m *Message) {
	c.replyText = m.Header.Get("Reply-Text")
	c.jobUUID = m.Header.Get("Job-UUID")
	c.c <- struct{}{}
}

func (c *CommandReply) wait() {
	<-c.c
}

func (c *CommandReply) Succeed() bool {
	i := strings.IndexByte(c.replyText, ' ')
	if i == -1 {
		return false
	}
	switch c.replyText[:i] {
	case replyOK:
		return true
	case replyERR:
		return false
	default:
		return false
	}
}

func (c *CommandReply) JobID() string {
	return c.jobUUID
}

func (c *CommandReply) Err() error {
	return c.err
}
