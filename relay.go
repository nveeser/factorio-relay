package main

import (
	"factorio-relay/udp"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

var (
	verbose = flag.Bool("verbose", false, "enable verbose logging")
	config  = newRelayConfigFlag("relays", "Comma separated list of port and destinations to broadcast to. "+
		"Example `3333=10.50.255.255,10.0.10.25;4444=192.168.0.255`")
)

func main() {
	for _, a := range os.Args {
		fmt.Printf("Command: %s\n", a)
	}
	flag.Parse()

	printInterfaces()

	fmt.Printf("Relay Config\n")
	for port, dstList := range *config {
		fmt.Printf("\tPort[%d]\n", port)
		for _, ip := range dstList {
			fmt.Printf("\t - Dest %s\n", ip)
		}
	}

	if err := relay(*config); err != nil {
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

func relay(config relayConfig) error {
	// Recover from the panic and print the stack trace.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
		}
	}()

	reader, err := udp.Listen("ip4:17", "0.0.0.0")
	if err != nil {
		return err
	}
	defer reader.Close()
	for {
		buf := make([]byte, 1024)
		packet, err := reader.ReadPacket(buf)
		if err != nil {
			return fmt.Errorf("error ReadPacket(): %w", err)
		}

		port := packet.UDPHeader.DstPort

		addrs, ok := config[port]
		if !ok {
			fmt.Printf("Port[%d] No relay configured\n", port)
			continue
		}
		for _, addr := range addrs {
			if *verbose {
				fmt.Printf("Port[%d] => %s\n", port, addr)
				fmt.Printf("Port[%d] IP: %s\n", port, packet.IPHeader)
				fmt.Printf("Port[%d] UDP: %s\n", port, packet.UDPHeader)
				fmt.Printf("Port[%d] Data: %v\n", port, packet.Payload)
			}
			packet.IPHeader.Dst = addr
			packet.UDPHeader.Checksum, err = udp.Checksum(packet)
			if err != nil {
				return fmt.Errorf("error UDPChecksum(): %w", err)
			}
			if err := reader.WritePacket(packet); err != nil {
				return fmt.Errorf("error WritePacket(): %w", err)
			}
		}
	}
}

func newRelayConfigFlag(name, desc string) *relayConfig {
	var config relayConfig = make(map[uint16][]net.IP)
	flag.Var(&config, name, desc)
	return &config
}

type relayConfig map[uint16][]net.IP

func (a *relayConfig) String() string {
	return fmt.Sprintf("%+v", *a)
}

func (a *relayConfig) Set(s string) error {
	if strings.Index(s, "=") == -1 {
		if ss, ok := os.LookupEnv(s); ok {
			fmt.Printf("Setting relays from environment %q\n", s)
			s = ss
		}
	}

	for _, v := range strings.Split(s, ";") {
		parts := strings.Split(v, "=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid relay specification: %s", v)
		}
		port, err := strconv.ParseUint(parts[0], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid relay port specification: %s (%s)", v, parts[0])
		}
		var ipset []net.IP
		for _, ipstr := range strings.Split(parts[1], ",") {
			ip := net.ParseIP(ipstr)
			if ip == nil {
				return fmt.Errorf("invalid address: %s", v)
			}
			ipset = append(ipset, ip)
		}
		(*a)[uint16(port)] = ipset
	}
	return nil
}
