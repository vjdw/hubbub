package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
	"time"

	"../network"
)

// Run the hubbub server
func Run() {
	bytesPerSample := 2 // int16
	channels := 2
	rate := 48000

	clientHostname := waitForClient()
	time.Sleep(2 * time.Second)

	pipeFile := "/tmp/pulsefifo"
	fmt.Println("open a named pipe file for read.")
	file, err := os.OpenFile(pipeFile, os.O_RDONLY|syscall.O_NONBLOCK, os.ModeNamedPipe)
	if err != nil {
		log.Fatal("Open named pipe file error:", err)
	}
	reader := bufio.NewReader(file)

	var packetSize = 32768
	samplesPerPacket := float32(packetSize / (bytesPerSample * channels))
	sampleDuration := 1.0 / float32(rate)
	packetAudioDuration := time.Duration(int32(1000000.0*samplesPerPacket*sampleDuration)) * time.Microsecond
	ticker := time.NewTicker(packetAudioDuration)
	go func() {
		for t := range ticker.C {
			data := make([]byte, packetSize)
			for i := range data {
				b, err := reader.ReadByte()
				if err != nil {
					fmt.Println("err getting byte")
				}
				data[i] = b
			}
			dataCopy := make([]byte, packetSize)
			copy(dataCopy, data)
			go network.Send(clientHostname, 1234, dataCopy)
			fmt.Println("Sent", packetSize, "bytes to", clientHostname, ":1234 at", t.Format(time.RFC3339))
		}
	}()

	for {
		time.Sleep(1 * time.Second)
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
