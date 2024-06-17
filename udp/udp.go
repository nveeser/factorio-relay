// Package udp provides a file types for reading UDP packets and spoofing them
// back onto the network.
package udp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/net/ipv4"
	"net"
)

// Conn wraps an existing ipv4.RawConn and parse the bytes into a package that
// can be modified and written back out.
type Conn struct {
	*ipv4.RawConn
}

// Listen opens a raw socket using net.ListenPacket and ipv4.NewRawConn
// to allow parsing the IP header, UDP Header and the payload.
func Listen(network, address string) (*Conn, error) {
	packetConn, err := net.ListenPacket(network, address)
	if err != nil {
		return nil, fmt.Errorf("error net.ListenPacket(%s): %s\n", network, err)
	}
	rawConn, err := ipv4.NewRawConn(packetConn)
	if err != nil {
		return nil, fmt.Errorf("error ipv4.NewRawConn(%s): %s\n", network, err)
	}
	return &Conn{
		RawConn: rawConn,
	}, nil
}

// ReadPacket reads the packet using the specified byte slice.
func (c *Conn) ReadPacket(b []byte) (*Packet, error) {
	hdr, ipPayload, _, err := c.RawConn.ReadFrom(b)
	if err != nil {
		return nil, fmt.Errorf("error during read: %w", err)
	}
	udpHdr := &Header{}
	uhd, payload := ipPayload[:HeaderLen], ipPayload[HeaderLen:]
	if err := udpHdr.ParseUDP(uhd); err != nil {
		fmt.Printf("error ParseUDP(): %s\n", err)
		return nil, fmt.Errorf("error error ParseUDP: %w", err)
	}
	return &Packet{
		IPHeader:  hdr,
		UDPHeader: udpHdr,
		Payload:   payload,
	}, nil
}

func (c *Conn) WritePacket(p *Packet) error {
	udpData, err := p.UDPHeader.Marshal()
	if err != nil {
		return fmt.Errorf("error udpHdr.Marshal(): %w", err)
	}
	ipPayload := append(udpData, p.Payload...)
	if err := c.RawConn.WriteTo(p.IPHeader, ipPayload, nil); err != nil {
		return fmt.Errorf("error wConn.WriteTo(): %w", err)
	}
	return nil
}

type Packet struct {
	IPHeader  *ipv4.Header
	UDPHeader *Header
	Payload   []byte
}

type Header struct {
	SrcPort  uint16
	DstPort  uint16
	Len      uint16
	Checksum uint16
}

const HeaderLen = 8

func (h *Header) String() string {
	if h == nil {
		return "<nil>"
	}
	return fmt.Sprintf("src=%d dst=%d len=%d cksum=%#x", h.SrcPort, h.DstPort, h.Len, h.Checksum)
}

func (h *Header) Marshal() ([]byte, error) {
	if h == nil {
		return nil, fmt.Errorf("nil header")
	}
	if h.Len < HeaderLen {
		return nil, fmt.Errorf("header too short")
	}
	b := make([]byte, HeaderLen)
	binary.BigEndian.PutUint16(b[0:2], h.SrcPort)
	binary.BigEndian.PutUint16(b[2:4], h.DstPort)
	binary.BigEndian.PutUint16(b[4:6], h.Len)
	binary.BigEndian.PutUint16(b[6:8], h.Checksum)
	return b, nil
}

func (h *Header) ParseUDP(b []byte) error {
	if h == nil {
		return fmt.Errorf("nil header")
	}
	if len(b) < HeaderLen {
		return fmt.Errorf("header too short")
	}
	h.SrcPort = binary.BigEndian.Uint16(b[0:2])
	h.DstPort = binary.BigEndian.Uint16(b[2:4])
	h.Len = binary.BigEndian.Uint16(b[4:6])
	h.Checksum = binary.BigEndian.Uint16(b[6:8])
	return nil
}

func Checksum(p *Packet) (uint16, error) {
	type pseudohdr struct {
		ipsrc   [4]byte
		ipdst   [4]byte
		zero    uint8
		ipproto uint8
		plen    uint16
		src     uint16
		dst     uint16
		ulen    uint16
		csum    uint16
	}
	phdr := &pseudohdr{
		zero:    0,
		ipproto: uint8(p.IPHeader.Protocol),
		plen:    p.UDPHeader.Len,
		src:     p.UDPHeader.SrcPort,
		dst:     p.UDPHeader.DstPort,
		ulen:    p.UDPHeader.Len,
		csum:    0,
	}
	if ip := p.IPHeader.Src.To4(); ip != nil {
		copy(phdr.ipsrc[:4], ip[:net.IPv4len])
	}
	if ip := p.IPHeader.Dst.To4(); ip != nil {
		copy(phdr.ipdst[:4], ip[:net.IPv4len])
	}
	var b bytes.Buffer
	if err := binary.Write(&b, binary.BigEndian, phdr); err != nil {
		return 0, err
	}
	b.Write(p.Payload)
	return checksum(b.Bytes()), nil
}

func checksum(buf []byte) uint16 {
	sum := uint32(0)

	for ; len(buf) >= 2; buf = buf[2:] {
		sum += uint32(buf[0])<<8 | uint32(buf[1])
	}
	if len(buf) > 0 {
		sum += uint32(buf[0]) << 8
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	csum := ^uint16(sum)
	/*
	 * From RFC 768:
	 * If the computed checksum is zero, it is transmitted as all ones (the
	 * equivalent in one's complement arithmetic). An all zero transmitted
	 * checksum value means that the transmitter generated no checksum (for
	 * debugging or for higher level protocols that don't care).
	 */
	if csum == 0 {
		csum = 0xffff
	}
	return csum
}
