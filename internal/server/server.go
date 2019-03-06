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
	bytesPerSample := 2 // int16
	channels := 2
	rate := 48000

	clientHostname := waitForClient()
	time.Sleep(5 * time.Second)

	pipeFile := "/tmp/pulsefifo"
	fmt.Println("open a named pipe file for read.")
	file, err := os.OpenFile(pipeFile, os.O_CREATE, os.ModeNamedPipe)
	if err != nil {
		log.Fatal("Open named pipe file error:", err)
	}

	var packetSize = 32768
	reader := bufio.NewReader(file)
	for {
		data := make([]byte, packetSize)
		for i := range data {
			b, _ := reader.ReadByte()
			data[i] = b
		}
		dataCopy := make([]byte, packetSize)
		copy(dataCopy, data)
		go network.Send(clientHostname, 1234, dataCopy)

		// Wouldn't need this if ReadByte would wait for a byte to be available
		samplesPerPacket := float32(packetSize / (bytesPerSample * channels))
		sampleDuration := 1.0 / float32(rate)
		packetAudioDuration := time.Duration(int32(1000000.0*samplesPerPacket*sampleDuration)) * time.Microsecond
		time.Sleep(packetAudioDuration)
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
