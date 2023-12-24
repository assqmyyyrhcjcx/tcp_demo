package iface

import "net"

type ITCPConn interface {
	Serve() error
	ReadPacket() (IPacket, error)
	WritePacket(IPacket) error
	WritePacketWithTimeout(IPacket, int) error
	GetLocalAddr() net.Addr
	GetLocalIPAddress() string
	GetRemoteAddr() net.Addr
	GetRemoteIPAddress() string
}
