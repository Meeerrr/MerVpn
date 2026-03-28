package main

import (
	"fmt"

	"mervpn/pkg/tunnel"
)

func main() {
	fmt.Println("======================================")
	fmt.Println("    MerVPN Engine Booting  ")
	fmt.Println("======================================")

	// Trigger the public Start() function from the tunnel package
	tunnel.Start()
}
