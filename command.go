package esl

import (
	"strconv"
	"sync"
)

var commandPool sync.Pool

var (
	strSendMsg = []byte("sendmsg")
	strApi     = []byte("api")
	strBgapi   = []byte("bgapi")
	strEvent   = []byte("event")
	strLFLF    = []byte("\n\n")
)

const (
	msgCommand = "call-command"
	msgApp     = "execute-app-name"
	msgArg     = "execute-app-arg"
	msgLoop    = "loops"
)

type CommandType uint8

const (
	MessageType CommandType = 1 + iota
	ApiType
	BgapiType
	EventType
)

func (c CommandType) String() string {
	switch c {
	case MessageType:
		return "message"
	case ApiType:
		return "api"
	case BgapiType:
		return "bgapi"
	case EventType:
		return "event"
	default:
		return "unknown"
	}
}

// https://freeswitch.org/confluence/display/FREESWITCH/mod_event_socket#mod_event_socket-3.CommandDocumentation
type Command struct {
	ct   CommandType
	uuid []byte
	kvs  Args
	buf  []byte
}

func AcquireCommand(t CommandType) *Command {
	got := commandPool.Get()
	if got == nil {
		return &Command{ct: t}
	}

	cmd := got.(*Command)
	cmd.ct = t
	return cmd
}

func releaseCommand(c *Command) {
	c.reset()
	commandPool.Put(c)
}

func (c *Command) reset() {
	c.uuid = c.uuid[:0]
	c.kvs.reset()
}

// Header return message headers
func (c *Command) Header() *Args {
	return &c.kvs
}

// api generate the api command
// https://freeswitch.org/confluence/display/FREESWITCH/mod_event_socket#mod_event_socket-3.1api
func (c *Command) api() []byte {
	var dst []byte

	dst = append(c.buf[:0], strApi...)
	dst = append(dst, ' ')
	dst = append(dst, c.kvs.GetBytes([]byte(msgApp))...)
	dst = append(dst, ' ')
	dst = append(dst, c.kvs.GetBytes([]byte(msgArg))...)
	dst = append(dst, strLFLF...)

	c.buf = dst
	return c.buf
}

// bgapi generate the bgapi command
// https://freeswitch.org/confluence/display/FREESWITCH/mod_event_socket#mod_event_socket-3.2bgapi
func (c *Command) bgapi() []byte {
	var dst []byte

	dst = append(c.buf[:0], strBgapi...)
	dst = append(dst, ' ')
	dst = append(dst, c.kvs.GetBytes([]byte(msgApp))...)
	dst = append(dst, ' ')
	dst = append(dst, c.kvs.GetBytes([]byte(msgArg))...)
	dst = append(dst, strLFLF...)

	c.buf = dst
	return c.buf
}

// https://freeswitch.org/confluence/display/FREESWITCH/mod_event_socket#mod_event_socket-3.5event
func (c *Command) event() []byte {
	var dst []byte

	dst = append(c.buf[:0], strEvent...)
	dst = append(dst, ' ')
	dst = append(dst, c.kvs.GetBytes([]byte(msgArg))...)
	dst = append(dst, strLFLF...)

	c.buf = dst
	return c.buf
}

// TODO: send event
// https://freeswitch.org/confluence/display/FREESWITCH/mod_event_socket#mod_event_socket-3.8sendevent
func (c *Command) sendEvent() []byte {
	return c.buf
}

// message generate message
// https://freeswitch.org/confluence/display/FREESWITCH/mod_event_socket#mod_event_socket-3.9sendmsg
// Outbound doesn't need uuid
func (c *Command) message() []byte {
	var dst []byte

	dst = append(c.buf[:0], strSendMsg...)
	if len(c.uuid) > 0 {
		dst = append(dst, ' ')
		dst = append(dst, c.uuid...)
	}
	dst = append(dst, '\n')
	dst = c.kvs.AppendBytes(dst)

	c.buf = dst
	return c.buf
}

func (c *Command) Bytes() []byte {
	switch c.ct {
	case MessageType:
		return c.message()
	case ApiType:
		return c.api()
	case BgapiType:
		return c.bgapi()
	case EventType:
		return c.event()
	default:
		return nil
	}
}

// SetUUID set uuid
// Outbound doesn't need uuid
func (c *Command) SetUUID(id string) *Command {
	c.uuid = append(c.uuid[:0], id...)
	return c
}

// SetCommand set command
// available command:
//   execute: invoke dialplan applications
//   hangup: hang up the call
//   unicast: hook up mod_spandsp for faxing over a socket
//   nomedia:
//   xferext:
func (c *Command) SetCommand(cmd string) *Command {
	c.kvs.Add(msgCommand, cmd)
	return c
}

// SetApp set the dialplan application
func (c *Command) SetApp(app string) *Command {
	c.kvs.Add(msgApp, app)
	return c
}

// SetArg set the dialplan application data
// arg must be shorter than 2048 bytes
// TODO: body support
func (c *Command) SetArg(arg string) *Command {
	if arg == "" {
		return c
	}
	c.kvs.Add(msgArg, arg)
	return c
}

// SetLoops set number of times to invoke the command, default: 1
func (c *Command) SetLoops(n int) *Command {
	c.kvs.Add(msgLoop, strconv.Itoa(n))
	return c
}

// SetHeader set custom header
func (c *Command) SetHeader(key, value string) *Command {
	c.kvs.Add(key, value)
	return c
}
