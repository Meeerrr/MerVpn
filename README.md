# MerVPN: Bare-Metal Layer 3 VPN Engine

MerVPN is a custom, full-duplex Virtual Private Network engine written entirely in Go. It operates directly at the network layer, bypassing high-level protocols to manually allocate virtual hardware, intercept raw IPv4 packets, and securely route them across a custom UDP bridge.

## Core Architecture

This engine is built on the principles of secure, high-speed concurrency and strict separation of concerns:

* **Bare-Metal TUN Allocation:** Interacts directly with the Linux kernel to spawn virtual network interfaces (`tun0`, `tun1`).
* **Raw Packet Dissection:** Utilizes C-style bitwise shifting (`>> 4`) to parse raw memory buffers and extract IPv4 routing headers with zero overhead.
* **Full-Duplex UDP Bridge:** Implements asynchronous Go routines to simultaneously listen for incoming network traffic while actively injecting intercepted packets into the OS.
* **Cryptography:** Secures the UDP tunnel using AES-256-GCM, providing both absolute confidentiality and tamper-proof cryptographic seals (Nonce-based authentication).
* **Sterile Environment Variables:** Enforces strict security by requiring 32-byte cryptographic keys to be injected via OS-level memory, keeping the source code completely free of hardcoded secrets.

## Prerequisites

To compile and run the MerVPN engine, the host environment must meet the following strict requirements:
* **Operating System:** Linux Kernel (Ubuntu, Debian, CentOS, etc.) or Windows Subsystem for Linux (WSL). Windows NT and macOS Darwin kernels require separate native networking drivers.
* **Compiler:** Go 1.21 or higher.
* **Permissions:** Execution strictly requires Root (`sudo` or `CAP_NET_ADMIN`) privileges to safely allocate and manipulate kernel-level network interfaces.
* **Environment:** A 32-character string must be present in the `MERVPN_KEY` environment variable prior to runtime to initialize the AES-GCM cipher block.

## Active Development Roadmap

MerVPN is currently transitioning from a point-to-point secure bridge into a fully-fledged commercial-grade internet gateway. Upcoming architectural upgrades include:

* **Network Address Translation (NAT):** Implementing IP masquerading on the server-side to actively forward intercepted client packets to the physical internet (WAN) and route replies back through the tunnel.
* **Global Route Override:** Upgrading the client application to dynamically rewrite the host operating system's default routing table (`0.0.0.0/0`), seamlessly funneling all system-wide traffic through the encrypted tunnel.
* **Multi-Client State Management:** Replacing the single-client memory tracker with a thread-safe concurrency map, allowing the server daemon to securely authenticate and multiplex traffic from hundreds of unique endpoints simultaneously.
