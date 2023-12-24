package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"tcp/iface"
)

type Protocol struct {
	maxPacketSize uint32
}

var headLength = 20

func (p *Protocol) ReadPacket(reader io.Reader) (iface.IPacket, error) {
	return p.ReadPacketLimit(reader, p.maxPacketSize)
}

func (*Protocol) ReadPacketLimit(reader io.Reader, size uint32) (iface.IPacket, error) {
	head := make([]byte, headLength)

	_, err := io.ReadFull(reader, head)
	if err != nil {
		return nil, err
	}

	packetLength := binary.BigEndian.Uint32(head)
	if size != 0 && packetLength > size {
		return nil, fmt.Errorf("packet too large(%v > %v)", packetLength, size)
	}

	buf := make([]byte, packetLength)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}
	return NewPacket(buf[1:]), nil
}

func (*Protocol) WritePacket(writer io.Writer, p Packet) error {
	var buf bytes.Buffer
	head := make([]byte, 4)
	data := p.Bytes()
	binary.BigEndian.PutUint32(head, uint32(len(data)))
	binary.Write(&buf, binary.BigEndian, head)
	binary.Write(&buf, binary.BigEndian, data)
	_, err := writer.Write(buf.Bytes())
	return err
}

func (p *Protocol) SetMaxPacketSize(n uint32) {
	p.maxPacketSize = n
}
