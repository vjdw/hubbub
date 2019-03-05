package network

import (
	"fmt"
	"net"
)

// Send bytes to host
func Send(host string, port int, data []byte) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)

	conn.Write(data)
	conn.Close()
}
