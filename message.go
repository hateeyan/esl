package esl

import "sync"

var messagePool sync.Pool

type Message struct {
	Header Header
	body   []byte
}

func NewMessage() *Message {
	return &Message{}
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
	return e.body
}

func (e *Message) ContentType() string {
	return e.Header.Get("Content-Type")
}

func (e *Message) reset() {
	e.Header.reset()
	e.body = e.body[:0]
}
