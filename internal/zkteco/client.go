package zkteco

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// Client wraps a TCP connection to a ZKTeco device.
type Client struct {
	conn      net.Conn
	sessionID uint16
	host      string
	port      int
	timeout   time.Duration
}

// Connect establishes TCP connection and creates a session.
func Connect(host string, port int, timeout time.Duration) (*Client, error) {
	if port == 0 {
		port = DefaultPort
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("zkteco connect %s: %w", addr, err)
	}

	c := &Client{
		conn:    conn,
		host:    host,
		port:    port,
		timeout: timeout,
	}

	if err := c.createSession(); err != nil {
		conn.Close()
		return nil, err
	}

	return c, nil
}

func (c *Client) createSession() error {
	// Send CmdConnect with password (empty bytes).
	pkt := createPacket(CmdConnect, 0, nil)
	if err := c.sendPacket(pkt); err != nil {
		return fmt.Errorf("session send: %w", err)
	}

	resp, err := c.recvPacket()
	if err != nil {
		return fmt.Errorf("session recv: %w", err)
	}

	if resp.Command == CmdAckOK || resp.Command == CmdAckData || resp.Command == CmdAckUData {
		c.sessionID = resp.SessionID
		return nil
	}

	// Second attempt with session from error response.
	if resp.Command == CmdAckError {
		pkt2 := createPacket(CmdConnect, resp.SessionID, nil)
		if err := c.sendPacket(pkt2); err != nil {
			return err
		}
		resp2, err := c.recvPacket()
		if err != nil {
			return err
		}
		c.sessionID = resp2.SessionID
		return nil
	}

	return fmt.Errorf("unexpected session response: cmd=%d", resp.Command)
}

// Close closes the TCP connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Host returns the device host.
func (c *Client) Host() string { return c.host }

// Port returns the device port.
func (c *Client) Port() int { return c.port }

// Exec sends a command and returns the response packet.
func (c *Client) Exec(cmd uint16, data []byte) (*packet, error) {
	pkt := createPacket(cmd, c.sessionID, data)
	if err := c.sendPacket(pkt); err != nil {
		return nil, err
	}
	resp, err := c.recvPacket()
	if err != nil {
		return nil, err
	}
	if resp.Command == CmdAckError {
		return nil, fmt.Errorf("zkteco error: cmd=%d", cmd)
	}
	return resp, nil
}

// ExecData sends a command and reads a multi-packet data response.
// Returns the full concatenated payload (excluding 4-byte size prefixes).
func (c *Client) ExecData(cmd uint16, data []byte) ([]byte, error) {
	pkt := createPacket(cmd, c.sessionID, data)
	if err := c.sendPacket(pkt); err != nil {
		return nil, err
	}

	var result []byte
	for {
		resp, err := c.recvPacket()
		if err != nil {
			return nil, err
		}

		switch resp.Command {
		case CmdAckOK:
			return result, nil
		case CmdAckError:
			return nil, fmt.Errorf("zkteco error: cmd=%d", cmd)
		case CmdAckData, CmdAckUData:
			// First 4 bytes of payload are the size.
			if len(resp.Data) >= 4 {
				result = append(result, resp.Data[4:]...)
			} else {
				result = append(result, resp.Data...)
			}
		default:
			return result, nil
		}
	}
}

// ExecSimple sends a command and expects a simple ACK_OK response.
func (c *Client) ExecSimple(cmd uint16, data []byte) error {
	resp, err := c.Exec(cmd, data)
	if err != nil {
		return err
	}
	if resp.Command != CmdAckOK {
		return fmt.Errorf("expected ACK_OK, got cmd=%d", resp.Command)
	}
	return nil
}

func (c *Client) sendPacket(p *packet) error {
	c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	_, err := c.conn.Write(p.encode())
	return err
}

func (c *Client) recvPacket() (*packet, error) {
	c.conn.SetReadDeadline(time.Now().Add(c.timeout))

	// Read header.
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return nil, fmt.Errorf("recv header: %w", err)
	}

	pkt, err := decodePacket(header)
	if err != nil {
		return nil, err
	}

	// For data responses, read more data.
	if pkt.Command == CmdAckData || pkt.Command == CmdAckUData {
		// First 4 bytes = remaining payload size.
		buf := make([]byte, 65536)
		total := 0
		for {
			n, err := c.conn.Read(buf[total:])
			if n > 0 {
				total += n
				// If we have at least 4 bytes, check the size prefix.
				if total >= 4 {
					payloadLen := int(binary.LittleEndian.Uint32(buf[:4]))
					if total >= payloadLen+4 {
						pkt.Data = buf[4 : payloadLen+4]
						return pkt, nil
					}
				}
			}
			if err != nil {
				if total > 0 {
					pkt.Data = buf[4:total]
					return pkt, nil
				}
				return nil, fmt.Errorf("recv data: %w", err)
			}
		}
	}

	return pkt, nil
}
