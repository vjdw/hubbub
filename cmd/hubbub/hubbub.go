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
	err := client.Run(*flagHostname)
	if err != nil {
		fmt.Println(err.Error())
	}
}
