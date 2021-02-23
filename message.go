package esl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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

func ReleaseMessage(msg *Message) {
	messagePool.Put(msg)
}

func (m *Message) parse(r *bufio.Reader) error {
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			return err
		}

		i := bytes.IndexByte(line, ':')
		if i == -1 {
			continue
		}
		m.Header.Add(line[:i], line[i+2:len(line)-1])

		peek, _ := r.Peek(1)
		if bytes.Compare(peek, []byte{'\n'}) == 0 {
			_, _ = r.Discard(1)
			n, err := m.Header.ContentLength()
			if err != nil {
				logger.Printf("unable to parse Content-Length: %v", err)
			}
			// has body
			if n > 0 {
				if len(m.body) < n {
					m.body = make([]byte, n)
				}
				_, err = io.ReadFull(r, m.body[:n])
				if err != nil {
					return err
				}
			}
			break
		}
	}
	return nil
}

// Body return message body
func (m *Message) Body() []byte {
	n, _ := m.Header.ContentLength()
	return m.body[m.bs:n]
}

// Bytes return original message
func (m *Message) Bytes() []byte {
	n, _ := m.Header.ContentLength()
	return m.body[:n]
}

func (m *Message) ContentType() string {
	return m.Header.Get("Content-Type")
}

func (m *Message) payload() *Message {
	buf := m.Body()
	m.Header.args.reset()
	var bs int
	for {
		l := bytes.IndexByte(buf, '\n')
		if l == -1 {
			break
		}
		i := bytes.IndexByte(buf, ':')
		if i != -1 {
			m.Header.Add(buf[:i], buf[i+2:l])
		}

		if len(buf) >= l+1 {
			if buf[l+1] != '\n' {
				buf = buf[l+1:]
				bs += l + 1
				continue
			}

			m.bs = bs + l + 2
			break
		}
	}
	return m
}

func (m *Message) reset() {
	m.bs = 0
	m.Header.reset()
}

type CommandReply struct {
	*Message
	err error
}

// replyResult return command execute result
func replyResult(reply string) bool {
	return strings.HasPrefix(reply, replyOK)
}

// JobID return bgapi job id
func (c *CommandReply) JobID() string {
	return c.Header.Get("Job-UUID")
}

func (c *CommandReply) Err() error {
	if c.err != nil {
		return c.err
	}

	reply := c.Message.Header.Get("Reply-Text")
	if reply == "" {
		reply = string(c.Message.Body())
	}
	if !replyResult(reply) {
		return fmt.Errorf("esl: failed to execute command: %s", strings.TrimSpace(reply))
	}
	return nil
}
