package main

import (
	"flag"
	"fmt"
	"os"

	"mervpn/pkg/tunnel"
)

func main() {
	mode := flag.String("mode", "", "Start the engine as 'client' or 'server'")
	flag.Parse()

	fmt.Println("======================================")
	fmt.Println("    MerVPN Engine Booting  ")
	fmt.Println("======================================")

	// SECURITY: Read the cryptographic key from the Operating System's environment
	secretKey := os.Getenv("MERVPN_KEY")

	// AES-256 requires a mathematically perfect 32-byte key. If it's missing or the wrong size, abort.
	if len(secretKey) != 32 {
		fmt.Println("[Fatal Error] You must provide exactly a 32-character key via the MERVPN_KEY environment variable.")
		fmt.Println("Example: sudo MERVPN_KEY=\"12345678901234567890123456789012\" go run ./cmd/mervpn/main.go -mode server")
		os.Exit(1)
	}

	if *mode == "server" {
		fmt.Println("[System] Booting in SERVER mode...")
	} else if *mode == "client" {
		fmt.Println("[System] Booting in CLIENT mode...")
	} else {
		fmt.Println("[Error] You must specify a mode to boot the engine.")
		os.Exit(1)
	}

	// Pass both the mode and the secure key into the engine block
	tunnel.Start(*mode, []byte(secretKey))
}
