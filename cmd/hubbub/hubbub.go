package main

import (
	"flag"
	"fmt"

	"../../internal/client"
	"../../internal/server"
)

func main() {
	flagHostname := flag.String("h", "", "Hubbub server hostname")
	flagServerMode := flag.Bool("s", false, "Server mode")

	flag.Parse()

	if *flagServerMode {
		fmt.Printf("No hostname specified, running in server mode.\n")
		server.Run()
	}

	fmt.Printf("Server hostname: %s\n", *flagHostname)

	device, err := client.GetDevice()
	if err != nil {
		fmt.Printf("Couldn't get audio device")
		return
	}

	err = client.Run(*flagHostname, device)
	if err != nil {
		fmt.Println(err.Error())
	}
}
