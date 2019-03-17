package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	alsa "github.com/cocoonlife/goalsa"

	"../network"
)

func getDevice() (b *alsa.PlaybackDevice, err error) {
	bp := alsa.BufferParams{BufferFrames: 0, PeriodFrames: 0, Periods: 0}
	return alsa.NewPlaybackDevice("default", 2, alsa.FormatS16LE, 48000, bp)
}

// Run a hubbub client
func Run(server string) (err error) {
	bytesPerSample := 2 // int16
	channels := 2
	rate := 48000

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

	audioBuf := make([]byte, 0)
	var packetSize = 32768
	var targetBufSize = 900000
	targetBufSize = targetBufSize - (targetBufSize % packetSize)
	fmt.Println("targetBufSize ", targetBufSize)

	audioBuf = fillBuffer(audioBuf, targetBufSize, packetSize, ser)
	var audioBufRemaining = targetBufSize

	var first = true
	toPlaySize := targetBufSize / 2
	toPlaySize = toPlaySize - (toPlaySize % packetSize)

	//var packetSize = 32768
	samplesPerPacket := float32(packetSize / (bytesPerSample * channels))
	sampleDuration := 1.0 / float32(rate)
	packetAudioDuration := time.Duration(int32(1000000.0*samplesPerPacket*sampleDuration)) * time.Microsecond
	ticker := time.NewTicker(packetAudioDuration)
	go func() {
		for t := range ticker.C {
			fmt.Println("Play chunk at", t)

			toPlay := audioBuf[:toPlaySize]
			audioBuf = audioBuf[toPlaySize:]
			audioBufRemaining -= toPlaySize

			toPlayReader := bytes.NewReader(toPlay)
			toPlayVals := make([]int16, toPlaySize/bytesPerSample)
			for i := range toPlayVals {
				var val int16
				binary.Read(toPlayReader, binary.LittleEndian, &val)
				toPlayVals[i] = val
				//fmt.Println("sample val: ", val)
			}

			_, err = device.Write(toPlayVals)
			if err != nil {
				fmt.Println(err)
			}

			if first {
				first = false
				toPlaySize = targetBufSize/4 - (targetBufSize % packetSize)
				targetBufSize = 3 * targetBufSize / 4
				targetBufSize = targetBufSize - (targetBufSize % packetSize)
			}

			// xyzzy need to make fillBuffer async from device.Write
			var toAdd = targetBufSize - audioBufRemaining
			audioBuf = fillBuffer(audioBuf, toAdd, packetSize, ser)
			audioBufRemaining += toAdd
		}
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}

func fillBuffer(buf []byte, toAdd int, packetSize int, ser *net.UDPConn) []byte {
	fmt.Println("Buffering...")
	var added = 0
	for added <= toAdd {
		p := make([]byte, packetSize)
		_, _, err := ser.ReadFromUDP(p)
		if err != nil {
			fmt.Println("Some error  %v", err)
			continue
		}
		pCopy := make([]byte, packetSize)
		copy(pCopy, p)
		buf = append(buf, pCopy...)
		added += packetSize
	}
	return buf
}
