package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	alsa "github.com/cocoonlife/goalsa"

	"../network"
)

func getDevice() (b *alsa.PlaybackDevice, err error) {
	bp := alsa.BufferParams{BufferFrames: 0, PeriodFrames: 0, Periods: 0}
	return alsa.NewPlaybackDevice("default", 2, alsa.FormatS16LE, 48000, bp)
}

// Run a hubbub client
func Run(server string) (err error) {
	device, err := getDevice()
	if err != nil {
		fmt.Printf("Couldn't get audio device")
		return err
	}

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
	network.Send(server, 1235, []byte("register"))

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
