package ptp

import (
	"encoding/binary"
	"fmt"
)

// Message represents a network packet sent over P2P network
// There are two types of messages available: comm and lan
// comm messages are communication control messages and internal
// for p2p network
// lan messages came from TAP interface and just being delivered
// to the TAP interface on another end without any processing
// except encryption/decryption
// Type of the message is determined by MagicCookie
type Message struct {
	MagicCookie uint16 // Type of the message
	Length      uint16 // Length of the payload
	Payload     []byte // Payload
}

// Marshal will create a byte slice with message representation
// that is ready to be sent over network
func (m *Message) Marshal() ([]byte, error) {
	if m.MagicCookie != MagicCookieComm && m.MagicCookie != MagicCookieLan {
		return nil, fmt.Errorf("wrong magic cookie")
	}
	buf := make([]byte, 1500)
	binary.BigEndian.PutUint16(buf[0:2], m.MagicCookie)
	binary.BigEndian.PutUint16(buf[2:4], m.Length)
	copy(buf[4:], m.Payload)
	return buf, nil
}

// Unmarshal will create a message from byte slice that was received
// from P2P network
func (m *Message) Unmarshal(buf []byte) error {
	if len(buf) < 3 {
		return fmt.Errorf("message is too small")
	}
	mc := binary.BigEndian.Uint16(buf[0:2])
	if mc != MagicCookie && mc != MagicCookieLan && mc != MagicCookieComm {
		return fmt.Errorf("wrong magic cookie")
	}
	m.MagicCookie = mc
	m.Length = binary.BigEndian.Uint16(buf[2:4])
	copy(m.Payload, buf[4:])
	return nil
}
