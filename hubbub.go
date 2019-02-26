package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/mesilliac/pulse-simple"
)

func main() {
	flagHostname := flag.String("h", "", "Hubbub server hostname")
	flag.Parse()

	if len(*flagHostname) > 0 {
		fmt.Printf("Server hostname: %s\n", *flagHostname)
		runClient(*flagHostname)
	} else {
		fmt.Printf("No hostname specified, running in server mode.\n")
		runServer()
	}
}

func runClient(server string) {
	// Connect to pulseaudio.
	ss := pulse.SampleSpec{Format: pulse.SAMPLE_S16LE, Rate: 48000, Channels: 2}
	pb, err := pulse.Playback("Hubbub Client", "Hubbub Client", &ss)
	defer pb.Free()
	defer pb.Drain()
	if err != nil {
		fmt.Printf("Could not create playback stream: %s\n", err)
		return
	}

	// Open connection to Hubbub server.
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", server, 1234))
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	ser, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	send(server, 1235, []byte("register"))

	pulseChunk := make([]byte, 800000)

	fmt.Printf("Buffering...")
	for i := 0; i < 800; i++ {
		p := make([]byte, 1000)
		_, _, err := ser.ReadFromUDP(p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		pCopy := make([]byte, 1000)
		copy(pCopy, p)
		pulseChunk = append(pulseChunk, pCopy...)
	}

	fmt.Printf("Play chunk\n")
	toPlay := pulseChunk[:400000]
	pulseChunk = pulseChunk[400000:]
	go pb.Write(toPlay)

	for {
		fmt.Printf("Play chunk\n")
		toPlay := pulseChunk[:200000]
		pulseChunk = pulseChunk[200000:]
		go pb.Write(toPlay)

		for i := 0; i < 200; i++ {
			p := make([]byte, 1000)
			_, _, err := ser.ReadFromUDP(p)
			if err != nil {
				fmt.Printf("Some error  %v", err)
				continue
			}
			pCopy := make([]byte, 1000)
			copy(pCopy, p)
			pulseChunk = append(pulseChunk, pCopy...)
		}
	}
}

func runServer() {
	client := waitForClient()

	var pipeFile = "/tmp/pulsefifo"
	fmt.Println("open a named pipe file for read.")
	file, err := os.OpenFile(pipeFile, os.O_CREATE, os.ModeNamedPipe)
	if err != nil {
		log.Fatal("Open named pipe file error:", err)
	}

	reader := bufio.NewReader(file)
	for {
		data := make([]byte, 1000)
		for i := range data {
			b, _ := reader.ReadByte()
			data[i] = b
		}
		dataCopy := make([]byte, 1000)
		copy(dataCopy, data)
		go send(client, 1234, dataCopy)
		time.Sleep((1000000 / 192) * time.Microsecond)
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
	client := remoteaddr.IP.String()

	return client
}

func send(host string, port int, data []byte) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)

	conn.Write(data)
	conn.Close()
}
