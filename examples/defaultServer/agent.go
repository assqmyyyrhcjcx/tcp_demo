package main

import (
	"log"
	"tcp/iface"
)

type agent struct{}

func (c *agent) OnMessage(conn iface.ITCPConn, p iface.IPacket) {
	log.Println("server receive:", string(p.Bytes()[1:]))
	//conn.WritePacket(&network.DefaultPacket{Type: 1, Body: []byte("yes")})
}

func (c *agent) OnConnected(conn iface.ITCPConn) {
	log.Println("new conn:", conn.GetRemoteAddr().String())
}

func (c *agent) OnDisconnected(conn iface.ITCPConn) {
	log.Printf("%s disconnected\n", conn.GetRemoteIPAddress())
}

func (c *agent) OnError(err error) {
	log.Println(err)
}
