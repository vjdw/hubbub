package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"../network"
)

// Run the hubbub server
func Run() {
	clientHostname := waitForClient()
	time.Sleep(5 * time.Second)

	var pipeFile = "/tmp/pulsefifo"
	fmt.Println("open a named pipe file for read.")
	file, err := os.OpenFile(pipeFile, os.O_CREATE, os.ModeNamedPipe)
	if err != nil {
		log.Fatal("Open named pipe file error:", err)
	}

	reader := bufio.NewReader(file)
	for {
		data := make([]byte, 1024)
		for i := range data {
			b, _ := reader.ReadByte()
			data[i] = b
		}
		dataCopy := make([]byte, 1024)
		copy(dataCopy, data)
		go network.Send(clientHostname, 1234, dataCopy)
		time.Sleep((1024000 / 192) * time.Microsecond)
	}
}

func waitForClient() string {
	fmt.Printf("Waiting for client to register.\n")
	addr, err := net.ResolveUDPAddr("udp", ":1235")
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return ""
	}
	ser, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return ""
	}
	p := make([]byte, 8)
	_, remoteaddr, err := ser.ReadFromUDP(p)
	fmt.Printf("%s request received from %v\n", p, remoteaddr)
	clientHostname := remoteaddr.IP.String()

	return clientHostname
}
