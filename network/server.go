package network

import (
	"errors"
	"log"
	"net"
	"os"
	"tcp/iface"
	"time"
)

const (
	//tcp conn 最大数据包大小
	defaultMaxPacketSize = 1024 << 10 //1MB

	readChanSize  = 100
	writeChanSize = 100
)

var (
	logger *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "", log.Lshortfile)
}

// TCPServer 结构定义
type TCPServer struct {
	//要监听的 TCP 地址
	tcpAddr string

	//listener
	listener *net.TCPListener

	//agent 是一个接口
	//它用于处理连接建立、关闭和数据接收
	agent    iface.IAgent
	protocol iface.IProtocol

	//如果 SRV 已关闭，请关闭用于通知所有会话退出的通道。
	exitChan chan struct{}

	readDeadline, writeDeadline time.Duration
	bucket                      *TCPConnBucket
}

// NewTCPServer 返回一个AsyncTCPServer实例
func NewTCPServer(tcpAddr string, agent iface.IAgent, protocol iface.IProtocol) *TCPServer {
	return &TCPServer{
		tcpAddr:  tcpAddr,
		agent:    agent,
		protocol: protocol,

		bucket:   NewTCPConnBucket(),
		exitChan: make(chan struct{}),
	}
}

// ListenAndServe 使用TCPServer的tcpAddr创建TCPListner并调用Server()方法开启监听
func (srv *TCPServer) ListenAndServe() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", srv.tcpAddr)
	if err != nil {
		return err
	}
	ln, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		return err
	}
	go srv.Serve(ln)
	return nil
}

// Serve 使用指定的TCPListener开启监听
func (srv *TCPServer) Serve(l *net.TCPListener) error {
	srv.listener = l
	defer func() {
		if r := recover(); r != nil {
			log.Println("Serve error", r)
		}
		srv.listener.Close()
	}()

	//清理无效连接
	go func() {
		for {
			srv.removeClosedTCPConn()
			time.Sleep(time.Millisecond * 10)
		}
	}()

	var tempDelay time.Duration
	for {
		select {
		case <-srv.exitChan:
			return errors.New("TCPServer Closed")
		default:
		}
		conn, err := srv.listener.AcceptTCP()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			log.Println("ln error:", err.Error())
			return err
		}
		tempDelay = 0
		tcpConn := srv.newTCPConn(conn, srv.agent, srv.protocol)
		tcpConn.setReadDeadline(srv.readDeadline)
		tcpConn.setWriteDeadline(srv.writeDeadline)
		srv.bucket.Put(tcpConn.GetRemoteAddr().String(), tcpConn)
	}
}

func (srv *TCPServer) newTCPConn(conn *net.TCPConn, agent iface.IAgent, protocol iface.IProtocol) *TCPConn {
	if agent == nil {
		agent = srv.agent
	}

	if protocol == nil {
		protocol = srv.protocol
	}

	c := NewTCPConn(conn, agent, protocol)
	c.Serve()
	return c
}

// Connect 使用指定的agent和protocol连接其他TCPServer，返回TCPConn
func (srv *TCPServer) Connect(ip string, agent iface.IAgent, protocol iface.IProtocol) (*TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	tcpConn := srv.newTCPConn(conn, agent, protocol)
	return tcpConn, nil

}

// Close 首先关闭所有连接，然后关闭TCPServer
func (srv *TCPServer) Close() {
	defer srv.listener.Close()
	for _, c := range srv.bucket.GetAll() {
		if !c.IsClosed() {
			c.Close()
		}
	}
}

func (srv *TCPServer) removeClosedTCPConn() {
	select {
	case <-srv.exitChan:
		return
	default:
		srv.bucket.removeClosedTCPConn()
	}
}

func (srv *TCPServer) GetAllTCPConn() []*TCPConn {
	conns := []*TCPConn{}
	for _, conn := range srv.bucket.GetAll() {
		conns = append(conns, conn)
	}
	return conns
}

func (srv *TCPServer) GetTCPConn(key string) *TCPConn {
	return srv.bucket.Get(key)
}

func (srv *TCPServer) SetReadDeadline(t time.Duration) {
	srv.readDeadline = t
}

func (srv *TCPServer) SetWriteDeadline(t time.Duration) {
	srv.writeDeadline = t
}
