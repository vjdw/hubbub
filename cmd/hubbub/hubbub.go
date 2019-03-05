package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/cocoonlife/goalsa"
)

func main() {

	flagHostname := flag.String("h", "", "Hubbub server hostname")
	flagServerMode := flag.Bool("s", false, "Server mode")
	flag.Parse()
	device, err := getDevice()
	if *flagServerMode {
		fmt.Printf("No hostname specified, running in server mode.\n")
		runServer()
	} // else if len(*flagHostname) > 0 {
	fmt.Printf("Server hostname: %s\n", *flagHostname)
	//device, err := getDevice()
	if err != nil {
		fmt.Printf("Couldn't get audio device")
	}
	err = runClient(*flagHostname, device)
	if err != nil {
		fmt.Println(err.Error())
	}
	//}
}

func getDevice() (b *alsa.PlaybackDevice, err error) {
	bp := alsa.BufferParams{BufferFrames: 0, PeriodFrames: 0, Periods: 0}
	return alsa.NewPlaybackDevice("default", 2, alsa.FormatS16LE, 48000, bp)
}

func runClient(server string, device *alsa.PlaybackDevice) (err error) {
	server = "localhost"
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

	pulseChunk := make([]byte, 0)

	fmt.Printf("Buffering...")
	for i := 0; i < 800; i++ {
		p := make([]byte, 1024)
		_, _, err := ser.ReadFromUDP(p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		pCopy := make([]byte, 1024)
		copy(pCopy, p)
		pulseChunk = append(pulseChunk, pCopy...)
	}

	toPlay := pulseChunk[:409600]
	toPlayReader := bytes.NewReader(toPlay)
	pulseChunk = pulseChunk[409600:]
	toPlayInts := make([]int16, 204800)
	for i := range toPlayInts {
		var num int16
		binary.Read(toPlayReader, binary.LittleEndian, &num)
		toPlayInts[i] = num
		fmt.Println("num: ", num)
	}
	_, err = device.Write(toPlayInts)
	if err != nil {
		return err
	}

	for {
		fmt.Printf("Play chunk\n")
		toPlay := pulseChunk[:102400]
		toPlayReader := bytes.NewReader(toPlay)
		pulseChunk = pulseChunk[102400:]
		toPlayInts := make([]int16, 50120)
		for i := range toPlayInts {
			var num int16
			binary.Read(toPlayReader, binary.LittleEndian, &num)
			toPlayInts[i] = num
			fmt.Println("num: ", num)
		}
		_, err := device.Write(toPlayInts)
		if err != nil {
			return err
		}

		for i := 0; i < 100; i++ {
			p := make([]byte, 1024)
			_, _, err := ser.ReadFromUDP(p)
			if err != nil {
				fmt.Printf("Some error  %v", err)
				continue
			}
			pCopy := make([]byte, 1024)
			copy(pCopy, p)
			pulseChunk = append(pulseChunk, pCopy...)
		}
	}
}

func runServer() {
	client := waitForClient()
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
		go send(client, 1234, dataCopy)
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
