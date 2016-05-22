package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/mesilliac/pulse-simple"
)

func main() {
	flagClient := flag.Bool("c", false, "Client mode.")
	flag.Parse()
	fmt.Printf("Client mode: ", *flagClient)

	if *flagClient {
		runClient()
	} else {
		recss := pulse.SampleSpec{pulse.SAMPLE_S16LE, 44100, 1}
		rec, err := pulse.Capture2("pulse-simple record", "pulse-simple record", &recss, "alsa_output.pci-0000_00_07.0.analog-stereo.monitor")
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
			go func() {
				sendToClient(&dataCopy)
			}()
		}

		// for {
		// 	data := make([]byte, 1*recss.Rate)
		// 	rec.Read(data)
		// 	sendToClient(&data)
		// }
	}
}

func sendToClient(data *[]byte) {
	//conn, err := net.Dial("udp", "llamatron:1236")
	addr, err := net.ResolveUDPAddr("udp", "llamatron:1234")
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)

	//fmt.Print(*data)
	conn.Write(*data)
	//fmt.Fprintf(conn, "Hi UDP client, How are you doing?")
	conn.Close()
}

func sendToServer(conn *net.UDPConn, addr *net.UDPAddr) {
	_, err := conn.WriteToUDP([]byte("From client: Hello I got your mesage "), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}

func runClient() {
	ss := pulse.SampleSpec{pulse.SAMPLE_S16LE, 44100, 1}
	pb, err := pulse.Playback("pulse-simple test", "playback test", &ss)
	defer pb.Free()
	defer pb.Drain()
	if err != nil {
		fmt.Printf("Could not create playback stream: %s\n", err)
		return
	}

	p := make([]byte, 1*ss.Rate)

	addr, err := net.ResolveUDPAddr("udp", ":1234")
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	// addr := net.UDPAddr{
	//     Port: 1236,
	//     IP: net.ParseIP("127.0.0.1"),
	// }

	ser, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	fmt.Printf("Started listening...")
	for {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		//fmt.Printf("Read a message from %v %s \n", remoteaddr, p)
		fmt.Printf("Read a message from %v\n", remoteaddr)

		pCopy := make([]byte, 1*ss.Rate)
		copy(pCopy, p)
		go func() {
			//go sendToServer(ser, remoteaddr)
			pb.Write(pCopy)
			return
		}()
	}
}
