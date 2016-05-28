package main

import (
	"flag"
	"fmt"
	"net"

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
	ss := pulse.SampleSpec{pulse.SAMPLE_S16LE, 44100, 1}
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

	fmt.Printf("Started listening...")
	p := make([]byte, 1*ss.Rate)
	for {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		fmt.Printf("Read a message from %v\n", remoteaddr)

		pCopy := make([]byte, 1*ss.Rate)
		copy(pCopy, p)
		go func() {
			pb.Write(pCopy)
			return
		}()
	}
}

func runServer() {
	client := waitForClient()

	// Connect to pulseaudio.
	recss := pulse.SampleSpec{pulse.SAMPLE_S16LE, 44100, 1}
	rec, err := pulse.Capture2("Hubbub Server", "Hubbub Server", &recss, "alsa_output.pci-0000_00_07.0.analog-stereo.monitor")
	defer rec.Free()
	defer rec.Drain()
	if err != nil {
		fmt.Printf("Could not create record stream: %s\n", err)
		return
	}

	for {
		data := make([]byte, 1*recss.Rate)
		rec.Read(data)

		dataCopy := make([]byte, 1*recss.Rate)
		copy(dataCopy, data)
		go send(client, 1234, dataCopy)
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
