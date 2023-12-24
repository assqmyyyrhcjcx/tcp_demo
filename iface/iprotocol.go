package iface

import "io"

type IProtocol interface {
	ReadPacket(reader io.Reader) (IPacket, error)
	WritePacket(writer io.Writer, msg IPacket) error
}
