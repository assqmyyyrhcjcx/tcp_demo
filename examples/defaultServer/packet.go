package main

import (
	"bytes"
	"encoding/binary"
)

type Packet struct {
	Body []byte
}

func NewPacket(body []byte) *Packet {
	return &Packet{
		Body: body,
	}
}

func (m *Packet) Bytes() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, m.Body)
	return buf.Bytes()
}
