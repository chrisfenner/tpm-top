package platform

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	powerOn    uint32 = 1
	powerOff   uint32 = 2
	nvOn       uint32 = 11
	nvOff      uint32 = 12
	sessionEnd uint32 = 20
)

// TcpConfig represents connection options for connecting to a running platform
// via TCP (e.g., the Microsoft reference TPM 2.0 simulator).
type TcpConfig struct {
	// Address is the full connection string for the running platform.
	Address string
}

// TcpPlatform is a connection to the running TCP platform.
type TcpPlatform struct {
	// conn is the open TCP connection to the running platform.
	conn net.Conn
}

// OpenTcpPlatform opens a connection to the running TCP platform.
func OpenTcpPlatform(c *TcpConfig) (*TcpPlatform, error) {
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		return nil, fmt.Errorf("could not dial TPM: %w", err)
	}
	return &TcpPlatform{
		conn: conn,
	}, nil
}

// sendCmd sends a command code to the running platform.
func (p TcpPlatform) sendCmd(cmd uint32) error {
	if err := binary.Write(p.conn, binary.BigEndian, cmd); err != nil {
		return fmt.Errorf("could not send platform command 0x%x: %w", cmd, err)
	}
	var rc uint32
	if err := binary.Read(p.conn, binary.BigEndian, &rc); err != nil {
		return fmt.Errorf("could not read platform response: %w", err)
	}
	if rc != 0 {
		return fmt.Errorf("error from TCP platform: 0x%x", rc)
	}
	return nil
}

// PowerOn powers on the platform.
func (p TcpPlatform) PowerOn() error {
	return p.sendCmd(powerOn)
}

// PowerOff powers off the platform.
func (p TcpPlatform) PowerOff() error {
	return p.sendCmd(powerOff)
}

// NVOn enables NV access.
func (p TcpPlatform) NVOn() error {
	return p.sendCmd(nvOn)
}

// NVOff disables NV access.
func (p TcpPlatform) NVOff() error {
	return p.sendCmd(nvOff)
}

// Close closes the connection to the running platform.
func (p TcpPlatform) Close() error {
	if err := binary.Write(p.conn, binary.BigEndian, sessionEnd); err != nil {
		p.conn.Close()
		return fmt.Errorf("error calling sessionEnd command on TCP TPM: %w", err)
	}
	return p.conn.Close()
}
