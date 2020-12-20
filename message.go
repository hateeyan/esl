package esl

import (
	"strings"
	"sync"
)

var messagePool sync.Pool

type Message struct {
	Header Header
	body   []byte
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
	return e.body[:n]
}

func (e *Message) ContentType() string {
	return e.Header.Get("Content-Type")
}

func (e *Message) reset() {
	e.Header.reset()
	e.body = e.body[:0]
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
