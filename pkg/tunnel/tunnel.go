package tunnel

import (
	"fmt"
	"log"
	"net"

	"github.com/songgao/water"
)

// Start allocates the virtual interface and establishes the full duplex UDP network bridge.
func Start(mode string) {
	// Allocate the virtual network interface from the Linux kernel
	config := water.Config{DeviceType: water.TUN}
	ifce, err := water.New(config)
	if err != nil {
		log.Fatalf("[Error] Failed to create TUN interface: %v", err)
	}
	fmt.Printf("[Tunnel] Virtual interface '%s' is online.\n", ifce.Name())

	var udpConn *net.UDPConn
	var serverConn *net.UDPConn

	// This variable will hold the Client's exact network location once they connect
	var activeClientAddr *net.UDPAddr

	if mode == "server" {
		addr, _ := net.ResolveUDPAddr("udp", ":9000")
		serverConn, err = net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatalf("[Error] Server failed to open port: %v", err)
		}
		fmt.Println("[Network] UDP Server actively listening on Port 9000...")

	} else if mode == "client" {
		addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
		udpConn, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			log.Fatalf("[Error] Client failed to dial server: %v", err)
		}
		fmt.Println("[Network] UDP Client locked onto Server at 127.0.0.1:9000...")
	}

	// Launch the background receiver to catch internet traffic and inject it into the OS
	go func() {
		networkBuffer := make([]byte, 1500)
		for {
			var n int
			var addr *net.UDPAddr
			var err error

			// Read incoming UDP packets based on our current mode
			if mode == "server" {
				n, addr, err = serverConn.ReadFromUDP(networkBuffer)
				// The Server saves the Client's return address to memory so it can reply later
				if err == nil {
					activeClientAddr = addr
				}
			} else {
				n, _, err = udpConn.ReadFromUDP(networkBuffer)
			}

			if err != nil {
				continue
			}

			// INJECTION: Force the received bytes directly into the local operating system
			ifce.Write(networkBuffer[:n])
			fmt.Printf("[Bridge -> OS] Injected %d bytes into the kernel.\n", n)
		}
	}()

	packet := make([]byte, 1500)

	// The main loop intercepts outgoing traffic from the OS and fires it over the bridge
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			continue
		}

		// ROUTING: Send the packet over the internet to the correct destination
		if mode == "client" {
			udpConn.Write(packet[:n])
			fmt.Println("[OS -> Bridge] Packet fired to Server.")

		} else if mode == "server" && activeClientAddr != nil {
			// The Server uses the return address it saved earlier to fire the reply
			serverConn.WriteToUDP(packet[:n], activeClientAddr)
			fmt.Println("[OS -> Bridge] Reply fired back to Client.")
		}
	}
}
