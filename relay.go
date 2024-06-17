package main

import (
	"factorio-relay/udp"
	"fmt"
	"golang.org/x/net/ipv4"
	"net"
	"runtime/debug"
)

var dests = flag.

func main() {
	// Recover from the panic and print the stack trace.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
		}
	}()

	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
				fmt.Printf("iface.Addrs() err: %s\n", err)
				continue
			}
			for _, addr := range addrs {
				fmt.Printf("%s: %s\n", iface.Name, addr)
			}
		}
	}
	reader, err := listen("ip4:17", "0.0.0.0")
	if err != nil {
		fmt.Printf("error net.ListenPacket(ip4:17): %s\n", err)
		return
	}
	writer, err := listen("ip4:17", "0.0.0.0")
	if err != nil {
		fmt.Printf("error net.ListenPacket(ip4:17): %s\n", err)
		return
	}
	for {
		buf := make([]byte, 1024)
		packet, err := reader.ReadPacket(buf)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		fmt.Printf("Header: %s", packet.IPHeader)
		fmt.Printf("UDP: %s\n", packet.UDPHeader)
		fmt.Printf("Data: %v\n", packet.Payload)

		packet.IPHeader.Dst = net.ParseIP("10.0.50.255")

		csum, err := udp.Checksum(packet)
		if err != nil {
			fmt.Printf("error UDPChecksum(): %s\n", err)
			return
		}
		packet.UDPHeader.Checksum = csum

		if err := writer.WritePacket(packet); err != nil {
			fmt.Printf("error UDPChecksum(): %s\n", err)
			return
		}
	}
}

func listen(network, address string) (*udp.Conn, error) {
	rc, err := net.ListenPacket(network, address)
	if err != nil {
		return nil, fmt.Errorf("error net.ListenPacket(%s): %s\n", network, err)
	}
	defer rc.Close()
	rConn, err := ipv4.NewRawConn(rc)
	if err != nil {
		return nil, fmt.Errorf("error ipv4.NewRawConn(%s): %s\n", network, err)
	}
	return &udp.Conn{RawConn: rConn}, nil
}
