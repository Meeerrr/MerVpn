package main

import (
	"flag"
	"fmt"
	"os"

	"mervpn/pkg/tunnel"
)

func main() {
	//  Define the flag.
	// We tell Go to look for a "-mode" flag, default it to empty (""), and provide a help description.
	mode := flag.String("mode", "", "Start the engine as 'client' or 'server'")

	// Parse the terminal command to extract the flags
	flag.Parse()

	fmt.Println("======================================")
	fmt.Println("    MerVPN Engine Booting  ")
	fmt.Println("======================================")

	// We use the pointer (*) to read the exact memory value of the flag
	if *mode == "server" {
		fmt.Println("[System] Booting in SERVER mode...")
		// Future: Server-specific UDP listening logic will go here

	} else if *mode == "client" {
		fmt.Println("[System] Booting in CLIENT mode...")
		// Future: Client-specific UDP sending logic will go here

	} else {

		// If the user forgets to type a mode, we shut down the engine safely instead of crashing.
		fmt.Println("[Error] You must specify a mode to boot the engine.")
		fmt.Println("Example: sudo go run ./cmd/mervpn/main.go -mode client")

		// os.Exit(1) tells the Linux operating system that the program closed due to an error.
		os.Exit(1)
	}

	// 5. Start the core interface (Both client and server need the TUN device)
	tunnel.Start(*mode)
}
