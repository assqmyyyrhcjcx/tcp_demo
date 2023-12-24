package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"tcp/iface"
	"tcp/network"
	"time"
)

func startClient(agent iface.IAgent, protocol iface.IProtocol) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:9001")
	if err != nil {
		panic(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(err)
	}
	tc := network.NewTCPConn(conn, agent, protocol)
	log.Println(tc.Serve())
	i := 0
	for {
		if tc.IsClosed() {
			break
		}
		msg := fmt.Sprintf("hello world%d", i)
		p := &Peoples{
			Nums: 10,
			Items: Items{
				Items: []Item{
					{
						Name: msg,
					},
				},
			},
		}
		output, err := xml.MarshalIndent(p, "  ", "    ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}

		//os.Stdout.Write([]byte(Header))
		//os.Stdout.Write(output)

		b := append([]byte(Header), output...)
		log.Println("client send: ", msg)
		tc.WritePacket(&network.DefaultPacket{Type: 1, Body: []byte(b)})
		i++
		time.Sleep(time.Second)
	}
}
