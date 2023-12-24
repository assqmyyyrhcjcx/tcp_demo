package main

import (
	"tcp/network"
)

func main() {
	protocol := &network.DefaultProtocol{}
	protocol.SetMaxPacketSize(1024)
	startServer(&agent{}, protocol)
	startClient(&agent{}, protocol)
	select {}
}
