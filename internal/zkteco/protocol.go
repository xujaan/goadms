package zkteco

import (
	"encoding/binary"
	"fmt"
)

// Protocol constants for ZKTeco TCP communication (port 4370).

const (
	DefaultPort = 4370

	// Commands
	CmdConnect    uint16 = 1000
	CmdExit       uint16 = 1001
	CmdRestart    uint16 = 1004
	CmdGetFreeSz  uint16 = 50
	CmdGetAttLog  uint16 = 201
	CmdGetUser    uint16 = 9
	CmdSetUser    uint16 = 8
	CmdDelUser    uint16 = 222

	// Responses
	CmdAckOK    uint16 = 2000
	CmdAckError uint16 = 2001
	CmdAckData  uint16 = 2002
	CmdAckUData uint16 = 2005

	headerSize = 8
)

// packet represents a ZKTeco protocol packet.
type packet struct {
	Command   uint16
	CheckSum  uint16
	SessionID uint16
	ReplyID   uint16
	Data      []byte
}

// encode serializes a packet to bytes for sending.
func (p *packet) encode() []byte {
	buf := make([]byte, headerSize+len(p.Data))
	binary.LittleEndian.PutUint16(buf[0:2], p.Command)
	binary.LittleEndian.PutUint16(buf[2:4], p.CheckSum)
	binary.LittleEndian.PutUint16(buf[4:6], p.SessionID)
	binary.LittleEndian.PutUint16(buf[6:8], p.ReplyID)
	copy(buf[8:], p.Data)
	return buf
}

// decodePacket parses bytes into a packet.
func decodePacket(data []byte) (*packet, error) {
	if len(data) < headerSize {
		return nil, fmt.Errorf("packet too short: %d bytes", len(data))
	}
	p := &packet{
		Command:   binary.LittleEndian.Uint16(data[0:2]),
		CheckSum:  binary.LittleEndian.Uint16(data[2:4]),
		SessionID: binary.LittleEndian.Uint16(data[4:6]),
		ReplyID:   binary.LittleEndian.Uint16(data[6:8]),
	}
	if len(data) > headerSize {
		p.Data = make([]byte, len(data)-headerSize)
		copy(p.Data, data[headerSize:])
	}
	return p, nil
}

// createPacket builds a command packet with computed checksum.
func createPacket(cmd, sessionID uint16, data []byte) *packet {
	p := &packet{
		Command:   cmd,
		SessionID: sessionID,
		Data:      data,
	}
	p.CheckSum = computeChecksum(p)
	return p
}

// computeChecksum calculates CRC16-CCITT over the packet contents.
// Checksum covers: command(2) + 0x0000(placeholder) + session(2) + reply(2) + data.
func computeChecksum(p *packet) uint16 {
	crc := crc16ccitt(nil, 0)
	crc = crc16ccitt([]byte{byte(p.Command), byte(p.Command >> 8)}, crc)
	crc = crc16ccitt([]byte{0, 0}, crc) // checksum placeholder
	crc = crc16ccitt([]byte{byte(p.SessionID), byte(p.SessionID >> 8)}, crc)
	crc = crc16ccitt([]byte{byte(p.ReplyID), byte(p.ReplyID >> 8)}, crc)
	crc = crc16ccitt(p.Data, crc)
	return crc
}

// crc16ccitt implements CRC16-CCITT (0x1021 polynomial).
func crc16ccitt(data []byte, seed uint16) uint16 {
	crc := seed
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
