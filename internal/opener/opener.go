package opener

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	sendCmd    uint32 = 8
	sessionEnd uint32 = 20
)

// TcpConfig represents connection options for connecting to a running TPM
// via TCP (e.g., the Microsoft reference TPM 2.0 simulator).
type TcpConfig struct {
	// Address is the full connection string for the running TPM.
	Address string
}

// tcpTpm represents a connection to a running TPM over TCP.
type tcpTpm struct {
	// conn is the open TCP connection to the running TPM.
	conn net.Conn
	// lastResp is the last response from the TPM.
	lastResp io.Reader
}

// OpenTcpTpm opens a connection to a running TPM via TCP (e.g., the Microsoft
// reference TPM 2.0 simulator).
func OpenTcpTpm(c *TcpConfig) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		return nil, fmt.Errorf("could not dial TPM: %w", err)
	}
	return &tcpTpm{
		conn: conn,
	}, nil
}

// Read reads the last response from the TCP TPM.
func (t *tcpTpm) Read(p []byte) (int, error) {
	return t.lastResp.Read(p)
}

// tcpCmdHdr represents a framed TCP TPM command header as defined in part D of
// https://trustedcomputinggroup.org/wp-content/uploads/TCG_TPM2_r1p59_Part4_SuppRoutines_code_pub.pdf
type tcpCmdHdr struct {
	tcpCmd   uint32
	locality uint8
	cmdLen   uint32
}

// Write frames the command and sends it to the TPM, immediately reading and
// caching the response for future calls to Read().
func (t *tcpTpm) Write(p []byte) (int, error) {
	cmd := tcpCmdHdr{
		tcpCmd:   sendCmd,
		locality: 0,
		cmdLen:   uint32(len(p)),
	}
	buf := bytes.Buffer{}
	if err := binary.Write(&buf, binary.BigEndian, cmd); err != nil {
		return 0, fmt.Errorf("could not frame TCP TPM command: %w", err)
	}
	if _, err := buf.Write(p); err != nil {
		return 0, fmt.Errorf("could not write command to buffer: %w", err)
	}
	if _, err := buf.WriteTo(t.conn); err != nil {
		return 0, fmt.Errorf("could not send TCP TPM command: %w", err)
	}

	var rspLen uint32
	if err := binary.Read(t.conn, binary.BigEndian, &rspLen); err != nil {
		return 0, fmt.Errorf("could not read TCP TPM response length: %w", err)
	}
	rsp := make([]byte, int(rspLen))
	if _, err := io.ReadFull(t.conn, rsp); err != nil {
		return 0, fmt.Errorf("could not read TCP TPM response: %w", err)
	}
	var rc uint32
	if err := binary.Read(t.conn, binary.BigEndian, &rc); err != nil {
		return 0, fmt.Errorf("could not read TCP TPM response code: %w", err)
	}
	if rc != 0 {
		return 0, fmt.Errorf("error from TCP TPM: 0x%x", rc)
	}
	t.lastResp = bytes.NewReader(rsp)
	return len(p), nil
}

// Close closes the connection to the TCP TPM.
func (t *tcpTpm) Close() error {
	if err := binary.Write(t.conn, binary.BigEndian, sessionEnd); err != nil {
		t.conn.Close()
		return fmt.Errorf("error calling sessionEnd command on TCP TPM: %w", err)
	}
	return t.conn.Close()
}
