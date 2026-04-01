package tunnel

import (
	"fmt"
	"log"
	"net"

	"github.com/songgao/water"
)

// Start allocates the TUN device and begins listening for raw IP packets.
func Start() {
	// Define the type of virtual hardware we want (TUN = Network Tunnel)
	config := water.Config{
		DeviceType: water.TUN,
	}

	// Ask the Operating System kernel to actually create the hardware
	ifce, err := water.New(config)
	if err != nil {
		log.Fatalf("[Error] Failed to create TUN interface. Details: %v", err)
	}

	fmt.Printf("[Tunnel] Success! Virtual interface '%s' is online.\n", ifce.Name())
	fmt.Println("[Tunnel] Listening for outbound IP packets...")

	// Optimized memory slice to hold incoming data
	packet := make([]byte, 1500)

	// Infinite Network Loop
	for {
		// Wait for the OS to hand us a packet
		n, err := ifce.Read(packet)
		if err != nil {
			continue // If an error occurs reading one packet, skip and wait for the next
		}

		// Pass only the exact bytes we received to the analyzer
		analyzePacket(packet[:n])
	}
}

// analyzePacket breaks down the raw binary data of an internet packet
func analyzePacket(packet []byte) {
	// A standard IPv4 header is 20 bytes long. Drop anything smaller.
	if len(packet) < 20 {
		return
	}

	// Bitwise shift to extract the IP version number from the first byte
	version := packet[0] >> 4

	if version == 4 {
		// In IPv4, bytes 16 through 19 always contain the target destination address
		destIP := net.IPv4(packet[16], packet[17], packet[18], packet[19])
		fmt.Printf("[Intercepted] IPv4 Packet attempting to reach -> %s (Size: %d bytes)\n", destIP.String(), len(packet))
	}
}
