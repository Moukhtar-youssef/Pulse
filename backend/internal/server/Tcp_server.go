package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/Moukhtar-youssef/Pulse/internal/protocol"
)

func StartTCPServer(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	fmt.Println("TCP server listening on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		go handleTCPClient(conn)
	}
}

func handleTCPClient(conn net.Conn) {
	defer conn.Close()

	for {
		header := make([]byte, 8)
		_, err := io.ReadFull(conn, header)
		if err != nil {
			fmt.Println("Client disconnected:", conn.RemoteAddr())
			return
		}

		magic := binary.BigEndian.Uint16(header[0:2])
		version := header[2]
		msgType := header[3]
		length := binary.BigEndian.Uint32(header[4:8])

		if magic != protocol.Magic {
			fmt.Println("invalid packet")
			return
		}

		payload := make([]byte, length)
		_, err = io.ReadFull(conn, payload)
		if err != nil {
			return
		}

		fmt.Println("Packet received")
		fmt.Println("Version:", version)
		fmt.Println("Type:", msgType)
		fmt.Println("Payload:", string(payload))
		fmt.Println("------------")
	}
}
