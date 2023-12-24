package network

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"tcp/iface"
	"time"
)

var (
	ErrConnClosing  = errors.New("use of closed network connection")
	ErrBufferFull   = errors.New("the async send buffer is full")
	ErrWriteTimeout = errors.New("async write packet timeout")
)

type TCPConn struct {
	agent    iface.IAgent
	protocol iface.IProtocol

	conn      *net.TCPConn
	readChan  chan iface.IPacket
	writeChan chan iface.IPacket

	readDeadline  time.Duration
	writeDeadline time.Duration

	exitChan  chan struct{}
	closeOnce sync.Once
	exitFlag  int32
	extraData map[string]interface{}
}

func NewTCPConn(conn *net.TCPConn, agent iface.IAgent, protocol iface.IProtocol) *TCPConn {
	c := &TCPConn{
		conn:     conn,
		agent:    agent,
		protocol: protocol,

		readChan:  make(chan iface.IPacket, readChanSize),
		writeChan: make(chan iface.IPacket, writeChanSize),

		exitChan: make(chan struct{}),
		exitFlag: 0,
	}
	return c
}

func (c *TCPConn) Serve() error {
	defer func() {
		if r := recover(); r != nil {
			logger.Println("tcp conn(%v) Serve error, %v ", c.GetRemoteIPAddress(), r)
		}
	}()
	if c.agent == nil || c.protocol == nil {
		err := fmt.Errorf("agent 和 protocol 不允许为 nil")
		c.Close()
		return err
	}
	atomic.StoreInt32(&c.exitFlag, 1)
	c.agent.OnConnected(c)
	go c.readLoop()
	go c.writeLoop()
	go c.handleLoop()
	return nil
}

func (c *TCPConn) readLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.exitChan:
			return
		default:
			if c.readDeadline > 0 {
				c.conn.SetReadDeadline(time.Now().Add(c.readDeadline))
			}
			p, err := c.protocol.ReadPacket(c.conn)
			if err != nil {
				if err != io.EOF {
					c.agent.OnError(err)
				}
				return
			}
			c.readChan <- p
		}
	}
}

func (c *TCPConn) ReadPacket() (iface.IPacket, error) {
	if c.protocol == nil {
		return nil, errors.New("no protocol impl")
	}
	return c.protocol.ReadPacket(c.conn)
}

func (c *TCPConn) writeLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for pkt := range c.writeChan {
		if pkt == nil {
			continue
		}
		if c.writeDeadline > 0 {
			c.conn.SetWriteDeadline(time.Now().Add(c.writeDeadline))
		}
		if err := c.protocol.WritePacket(c.conn, pkt); err != nil {
			c.agent.OnError(err)
			return
		}
	}
}

func (c *TCPConn) handleLoop() {
	defer func() {
		recover()
		c.Close()
	}()
	for p := range c.readChan {
		if p == nil {
			continue
		}
		c.agent.OnMessage(c, p)
	}
}

func (c *TCPConn) WritePacket(p iface.IPacket) error {
	if c.IsClosed() {
		return ErrConnClosing
	}
	select {
	case c.writeChan <- p:
		return nil
	default:
		return ErrBufferFull
	}
}

func (c *TCPConn) WritePacketWithTimeout(p iface.IPacket, sec int) error {
	if c.IsClosed() {
		return ErrConnClosing
	}
	select {
	case c.writeChan <- p:
		return nil
	case <-time.After(time.Second * time.Duration(sec)):
		return ErrWriteTimeout
	}
}

func (c *TCPConn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.exitFlag, 0)
		close(c.exitChan)
		close(c.writeChan)
		close(c.readChan)
		if c.agent != nil {
			c.agent.OnDisconnected(c)
		}
		c.conn.Close()
	})
}

func (c *TCPConn) GetRawConn() *net.TCPConn {
	return c.conn
}

func (c *TCPConn) IsClosed() bool {
	return atomic.LoadInt32(&c.exitFlag) == 0
}

func (c *TCPConn) GetLocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// LocalIPAddress 返回socket连接本地的ip地址
func (c *TCPConn) GetLocalIPAddress() string {
	return strings.Split(c.GetLocalAddr().String(), ":")[0]
}

func (c *TCPConn) GetRemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TCPConn) GetRemoteIPAddress() string {
	return strings.Split(c.GetRemoteAddr().String(), ":")[0]
}

func (c *TCPConn) setReadDeadline(t time.Duration) {
	c.readDeadline = t
}

func (c *TCPConn) setWriteDeadline(t time.Duration) {
	c.writeDeadline = t
}

func (c *TCPConn) SetExtraData(key string, data interface{}) {
	if c.extraData == nil {
		c.extraData = make(map[string]interface{})
	}
	c.extraData[key] = data
}

func (c *TCPConn) GetExtraData(key string) interface{} {
	if data, ok := c.extraData[key]; ok {
		return data
	}
	return nil
}
