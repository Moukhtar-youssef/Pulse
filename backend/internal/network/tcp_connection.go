package network

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/Moukhtar-youssef/Pulse/internal/protocol"
)

type TCPConnection struct {
	conn net.Conn
}

func NewTCPConnection() *TCPConnection {
	return &TCPConnection{}
}

func (t *TCPConnection) Connect(host string, port int) error {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	t.conn = conn
	return nil
}

func (t *TCPConnection) SendPacket(packetType uint8, payload string) error {
	header := protocol.Header{
		Magic:   protocol.Magic,
		Version: protocol.Version,
		Type:    packetType,
		Length:  uint32(len(payload)),
	}

	err := binary.Write(t.conn, binary.BigEndian, header)
	if err != nil {
		return err
	}

	_, err = t.conn.Write([]byte(payload))
	return err
}

func (t *TCPConnection) Close() {
	if t.conn != nil {
		t.conn.Close()
	}
}
