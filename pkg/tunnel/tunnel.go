package tunnel

import (
	"fmt"
	"log"
	"net"

	"github.com/songgao/water"
)

// Start allocates the virtual interface and establishes the UDP network bridge.
// It requires the 'mode' string to determine if it should act as a sender or receiver.
func Start(mode string) {
	// Ask the operating system kernel to allocate a virtual network tunnel interface
	config := water.Config{DeviceType: water.TUN}
	ifce, err := water.New(config)
	if err != nil {
		log.Fatalf("[Error] Failed to create TUN interface: %v", err)
	}
	fmt.Printf("[Tunnel] Virtual interface '%s' is online.\n", ifce.Name())

	// Prepare a pointer to hold our active network socket
	var udpConn *net.UDPConn

	if mode == "server" {
		// The Server acts as the receiver, opening a local port and waiting for incoming traffic
		addr, _ := net.ResolveUDPAddr("udp", ":9000")
		udpConn, err = net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatalf("[Error] Server failed to open port: %v", err)
		}
		fmt.Println("[Network] UDP Server actively listening on Port 9000...")

	} else if mode == "client" {
		// The Client acts as the sender, aiming its network socket directly at the Server's address
		// Note: Using the local loopback address (127.0.0.1) for testing purposes
		addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
		udpConn, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			log.Fatalf("[Error] Client failed to dial server: %v", err)
		}
		fmt.Println("[Network] UDP Client locked onto Server at 127.0.0.1:9000...")
	}

	// Launch a background Goroutine to continuously monitor the network socket for incoming data.
	// This runs concurrently alongside the main interface loop so the engine never blocks or freezes.
	go func() {
		networkBuffer := make([]byte, 1500)
		for {
			n, _, err := udpConn.ReadFromUDP(networkBuffer)
			if err != nil {
				continue
			}
			fmt.Printf("\n[Network Bridge] SUCCESS! Received %d raw bytes over UDP!\n", n)
		}
	}()

	// Create a memory buffer optimized for standard internet transmission unit sizes
	packet := make([]byte, 1500)

	// Enter the infinite loop to intercept packets routed into the virtual tunnel by the OS
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			continue
		}

		analyzePacket(packet[:n])

		// When operating as the client, immediately transmit intercepted packets across the UDP bridge
		if mode == "client" {
			udpConn.Write(packet[:n])
			fmt.Println("[Tunnel] -> Packet physically fired across UDP socket.")
		}
	}
}

// analyzePacket extracts routing metadata directly from raw binary memory
func analyzePacket(packet []byte) {
	// Ensure the packet contains at least a standard IPv4 header length
	if len(packet) < 20 {
		return
	}

	// Use a bitwise shift to isolate the IP protocol version from the first byte
	version := packet[0] >> 4
	if version == 4 {
		destIP := net.IPv4(packet[16], packet[17], packet[18], packet[19])
		fmt.Printf("[Intercepted] IPv4 Packet attempting to reach -> %s (Size: %d bytes)\n", destIP.String(), len(packet))
	}
}
