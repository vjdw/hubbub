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

	laddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", "0.0.0.0", 1234))
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	ser, err := net.ListenUDP("udp", laddr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	network.Send(server, 1235, []byte("register"))

	var pcmStream = make(chan []byte, 512)

	// Receive audio from network and put on pcmStream channel
	go func() {
		audioBuf := make([]byte, 0)
		var packetSize = 32768
		var targetBufSize = 300000
		targetBufSize = targetBufSize - (targetBufSize % packetSize)
		fmt.Println("targetBufSize ", targetBufSize)

		audioBufRemaining, audioBuf := fillBuffer(audioBuf, targetBufSize, packetSize, ser)
		fmt.Println("Added", audioBufRemaining, "bytes to buffer")

		var first = true
		toPlaySize := targetBufSize / 2
		toPlaySize = toPlaySize - (toPlaySize % packetSize)

		samplesPerPacket := float32(packetSize / (bytesPerSample * channels))
		sampleDuration := 1.0 / float32(rate)
		packetAudioDuration := time.Duration(int32(1000000.0*samplesPerPacket*sampleDuration)) * time.Microsecond
		ticker := time.NewTicker(packetAudioDuration)

		for {
			<-ticker.C
			var toAdd = targetBufSize - audioBufRemaining
			n := 0
			n, audioBuf = fillBuffer(audioBuf, toAdd, packetSize, ser)
			fmt.Println("Added", n, "bytes to buffer")
			audioBufRemaining += n

			toPlay := audioBuf[:toPlaySize]
			audioBuf = audioBuf[toPlaySize:]
			audioBufRemaining -= toPlaySize
			pcmStream <- toPlay

			if first {
				first = false
				toPlaySize = targetBufSize/4 - (targetBufSize % packetSize)
				targetBufSize = 3 * targetBufSize / 4
				targetBufSize = targetBufSize - (targetBufSize % packetSize)
			}
		}
	}()

	// Receive audio from pcmStream channel and write to audio device
	go func() {
		for {
			toPlay := <-pcmStream
			toPlaySize := len(toPlay)
			toPlayReader := bytes.NewReader(toPlay)
			toPlayVals := make([]int16, toPlaySize/bytesPerSample)
			for i := range toPlayVals {
				var val int16
				binary.Read(toPlayReader, binary.LittleEndian, &val)
				toPlayVals[i] = val
			}

			_, err = device.Write(toPlayVals)
			fmt.Println("Sent", toPlaySize, "bytes to device")
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}

func fillBuffer(buf []byte, toAdd int, packetSize int, ser *net.UDPConn) (int, []byte) {
	var added = 0
	for added <= toAdd {
		p := make([]byte, packetSize)
		n, _, err := ser.ReadFromUDP(p)
		if err != nil {
			fmt.Println("Error reading from UDP %v", err)
			continue
		}
		pCopy := make([]byte, n)
		copy(pCopy, p)
		buf = append(buf, pCopy...)
		added += n
		fmt.Printf("\r%d bytes added to buffer", added)
	}
	fmt.Printf("\n")
	return added, buf
}
