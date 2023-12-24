package main

import (
	"tcp/iface"
	"tcp/network"
)

func startServer(agent iface.IAgent, protocol iface.IProtocol) error {
	srv := network.NewTCPServer("localhost:9001", agent, protocol)
	return srv.ListenAndServe()
}
