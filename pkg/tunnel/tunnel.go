package tunnel

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/songgao/water"
)

// Start allocates the virtual interface and establishes the encrypted UDP bridge.
// It now requires the cryptographic key to be passed in, preventing hardcoded secrets.
func Start(mode string, key []byte) {
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
		addr, _ := net.ResolveUDPAddr("udp", "192.168.1.35:9000")
		udpConn, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			log.Fatalf("[Error] Client failed to dial server: %v", err)
		}
		fmt.Println("[Network] UDP Client locked onto Server at 127.0.0.1:9000...")
	}

	// Launch the background receiver to catch internet traffic and inject it into the OS
	go func() {
		networkBuffer := make([]byte, 2000) // Increased buffer size to handle encryption overhead
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

			// CRYPTO: Attempt to decrypt the incoming packet using the dynamic key
			decryptedPacket, err := decryptPacket(networkBuffer[:n], key)
			if err != nil {
				// If decryption fails (wrong key or tampered packet), silently drop it to prevent attacks
				fmt.Println("[Security Warning] Dropped invalid or tampered packet.")
				continue
			}

			// INJECTION: Force the clean, decrypted bytes directly into the local operating system
			ifce.Write(decryptedPacket)
			fmt.Printf("[Bridge -> OS] Decrypted and injected %d bytes.\n", len(decryptedPacket))
		}
	}()

	packet := make([]byte, 1500)

	// The main loop intercepts outgoing traffic from the OS and fires it over the bridge
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			continue
		}

		// Analyze the raw packet for logging purposes
		analyzePacket(packet[:n])

		// CRYPTO: Encrypt the packet before it touches the wild internet using the dynamic key
		encryptedPacket := encryptPacket(packet[:n], key)

		// ROUTING: Send the encrypted packet over the internet to the correct destination
		if mode == "client" {
			udpConn.Write(encryptedPacket)
			fmt.Println("[OS -> Bridge] Packet encrypted and fired to Server.")

		} else if mode == "server" && activeClientAddr != nil {
			// The Server uses the return address it saved earlier to fire the reply
			serverConn.WriteToUDP(encryptedPacket, activeClientAddr)
			fmt.Println("[OS -> Bridge] Encrypted reply fired back to Client.")
		}
	}
}

// encryptPacket scrambles the raw data and attaches a cryptographic seal using the provided dynamic key
func encryptPacket(plaintext []byte, key []byte) []byte {
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	// A Nonce (Number Used Once) ensures identical packets look completely different when encrypted
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)

	return gcm.Seal(nonce, nonce, plaintext, nil)
}

// decryptPacket verifies the cryptographic seal and restores the original data using the provided dynamic key
func decryptPacket(ciphertext []byte, key []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("packet too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, actualCiphertext, nil)
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
