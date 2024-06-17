package main

import (
	"factorio-relay/udp"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"strings"
)

var addrs addressList

var verbose = flag.Bool("verbose", false, "enable verbose logging")

func init() {
	flag.Var(&addrs, "networks", "List of networks to broadcast to")
}

func main() {
	flag.Parse()
	for _, a := range os.Args {
		fmt.Printf("Command: %s\n", a)
	}
	printInterfaces()
	if err := relay(); err != nil {
		fmt.Printf("Error during relay: %s", err)
	}
}
func printInterfaces() {
	list, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, iface := range list {
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

func relay() error {
	// Recover from the panic and print the stack trace.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
		}
	}()

	reader, err := udp.Listen("ip4:17", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("error net.ListenPacket(ip4:17): %w", err)
	}
	defer reader.Close()
	writer, err := udp.Listen("ip4:17", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("error net.ListenPacket(ip4:17): %w", err)
	}
	defer writer.Close()
	for {
		buf := make([]byte, 1024)
		packet, err := reader.ReadPacket(buf)
		if err != nil {
			return fmt.Errorf("error ReadPacket(): %w", err)
		}

		if *verbose {
			fmt.Printf("Header: %s", packet.IPHeader)
			fmt.Printf("UDP: %s\n", packet.UDPHeader)
			fmt.Printf("Data: %v\n", packet.Payload)
		}

		if len(addrs) == 0 {
			fmt.Printf("No Networks to relay to")
			continue
		}

		for _, addr := range addrs {
			if *verbose {
				fmt.Printf("Publish to %s", addr)
			}
			packet.IPHeader.Dst = addr
			packet.UDPHeader.Checksum, err = udp.Checksum(packet)
			if err != nil {
				return fmt.Errorf("error UDPChecksum(): %w", err)
			}
			if err := writer.WritePacket(packet); err != nil {
				return fmt.Errorf("error WritePacket(): %w", err)
			}
		}
	}
}

type addressList []net.IP

func (a *addressList) String() string {
	var s []string
	for _, ip := range *a {
		s = append(s, ip.String())
	}
	return strings.Join(s, ",")
}

func (a *addressList) Set(s string) error {
	vals := strings.Split(s, ",")
	for _, v := range vals {
		ip := net.ParseIP(v)
		if ip == nil {
			return fmt.Errorf("invalid address: %s", v)
		}
		*a = append(*a, ip)
	}
	return nil
}
