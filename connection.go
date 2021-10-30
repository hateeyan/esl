package esl

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
)

type Applications struct {
	// OnReconnect func called when reconnecting
	OnReconnect func(c *Inbound, err error)
	// OnEvent func called when an event message fetched
	OnEvent func(msg *Message)
}

type connectionType uint8

const (
	inbound connectionType = 1 + iota
	outbound
)

var connectionPool sync.Pool

type Connection struct {
	apps *Applications

	conn net.Conn
	ct   connectionType
	r    *bufio.Reader

	mu sync.Mutex

	channelData      *Message
	commandReplyChan chan *Message
}

func acquireConnection(conn net.Conn, t connectionType) *Connection {
	got := connectionPool.Get()
	if got == nil {
		return &Connection{
			conn:             conn,
			ct:               t,
			r:                bufio.NewReader(conn),
			commandReplyChan: make(chan *Message, 1),
		}
	}
	o := got.(*Connection)
	o.reset(conn)
	return o
}

func releaseOutbound(o *Connection) {
	connectionPool.Put(o)
}

func (c *Connection) reset(conn net.Conn) {
	c.conn = conn
	c.r.Reset(conn)
	c.channelData.reset()
}

func (c *Connection) waitMessage() {
	for {
		msg, err := parseMessage(c.r)
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Printf("unable to parse message: %v", err)
			break
		}
		ct := msg.ContentType()
		switch ct {
		case commandReply:
			c.produceReply(msg)
		case apiResponse:
			c.produceReply(msg)
		case eventPlain:
			if c.apps.OnEvent != nil {
				go func(msg *Message) {
					c.apps.OnEvent(msg.payload())
					ReleaseMessage(msg)
				}(msg)
			}
		case disconnectNotice:
			return
		default:
			logger.Printf("unhandled content type: %s", ct)
			ReleaseMessage(msg)
		}
	}
}

func parseMessage(r *bufio.Reader) (*Message, error) {
	e := acquireMessage()
	err := e.parse(r)
	return e, err
}

// Info CHANNEL_DATA event after connected
func (c *Connection) Info() *Message {
	return c.channelData
}

// connect is the first command to send to FreeSWITCH side
func (c *Connection) connect() error {
	_, err := c.conn.Write([]byte("connect\n\n"))
	if err != nil {
		return err
	}
	msg, err := parseMessage(c.r)
	if err != nil {
		return err
	}
	c.channelData = msg
	return nil
}

func (c *Connection) produceReply(msg *Message) {
	c.commandReplyChan <- msg
}

func (c *Connection) waitReply() *Message {
	return <-c.commandReplyChan
}

func (c *Connection) send(cmd *Command) error {
	buf := cmd.Bytes()
	if buf == nil {
		return fmt.Errorf("invalid command type: %s", cmd.ct)
	}
	_, err := c.conn.Write(buf)
	return err
}

func (c *Connection) Command(cmd *Command) (reply CommandReply) {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.send(cmd)
	if err != nil {
		reply.err = err
		return
	}

	reply.Message = c.waitReply()
	return
}

// Execute send a command to FreeSWITCH
func (c *Connection) Execute(app, arg string) (reply CommandReply) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cmd := AcquireCommand(MessageType).
		SetCommand("execute").
		SetApp(app).
		SetArg(arg)
	err := c.send(cmd)
	releaseCommand(cmd)
	if err != nil {
		reply.err = err
		return
	}

	reply.Message = c.waitReply()
	return
}

// Api send a FreeSWITCH API command, blocking mode
func (c *Connection) Api(api, arg string) (reply CommandReply) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cmd := AcquireCommand(ApiType).SetApp(api).SetArg(arg)
	err := c.send(cmd)
	releaseCommand(cmd)
	if err != nil {
		reply.err = err
		return
	}

	reply.Message = c.waitReply()
	return
}

// Bgapi send a FreeSWITCH API command, non-blocking mode
// return job id
func (c *Connection) Bgapi(app, arg string) (*Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cmd := AcquireCommand(BgapiType).SetApp(app).SetArg(arg)
	err := c.send(cmd)
	releaseCommand(cmd)
	if err != nil {
		return nil, err
	}

	msg := c.waitReply()
	return msg, nil
}

// Event send a FreeSWITCH event command
func (c *Connection) Event(arg string) (reply CommandReply) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cmd := AcquireCommand(EventType).SetArg(arg)
	err := c.send(cmd)
	releaseCommand(cmd)
	if err != nil {
		reply.err = err
		return
	}

	reply.Message = c.waitReply()
	return
}

// Hangup Hangs up a channel
// cause: https://freeswitch.org/confluence/display/FREESWITCH/Hangup+Cause+Code+Table
func (c *Connection) Hangup(cause string) (reply CommandReply) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cmd := AcquireCommand(MessageType).SetCommand("hangup")
	if cause != "" {
		cmd.SetHeader("hangup-cause", cause)
	}
	err := c.send(cmd)
	releaseCommand(cmd)
	if err != nil {
		reply.err = err
		return
	}

	reply.Message = c.waitReply()
	return
}

func (c *Connection) Close() error {
	return c.conn.Close()
}
