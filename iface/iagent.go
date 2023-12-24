package iface

// IAgent 用于连接的各种事件处理
type IAgent interface {
	// 链接建立
	OnConnected(conn ITCPConn)
	// 消息处理
	OnMessage(conn ITCPConn, p IPacket)
	// 链接断开
	OnDisconnected(conn ITCPConn)
	// 错误
	OnError(err error)
}
