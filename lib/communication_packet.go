package ptp

import "encoding/binary"

// CommunicationPacket is a message sent/received by communication subsystem
// Packet struct has common fields for all types of communication packets
// This means, that sometimes some fields might be empty if they are not used
// in particular subsystem that they intend too
type CommunicationPacket struct {
	PacketType uint16 // Type of the packet used to determine which subsystem supposed to receive it
	HashLength uint16 // Length of the hash string
	Hash       []byte // Unique ID of a swarm
	ID         []byte // Unique ID of a src/dst peer
	Payload    []byte // Extra payload
}

// Marshal will generate bytes slice from Communication Packet
// which can be used as a payload for a future network message
func (c *CommunicationPacket) Marshal() ([]byte, error) {
	buffer := make([]byte, 1200)
	binary.BigEndian.PutUint16(buffer[0:2], c.PacketType)
	binary.BigEndian.PutUint16(buffer[2:4], c.HashLength)
	hOffset := 4 + c.HashLength
	iOffset := hOffset + uint16(len(c.ID)) // len(ID) must always return 36
	copy(buffer[4:hOffset], c.Hash)
	copy(buffer[hOffset:iOffset], c.ID)
	copy(buffer[iOffset:], c.Payload)
	return buffer, nil
}

// Unmarshal accepts a bytes slice and fill CommunicationPacket fields
func (c *CommunicationPacket) Unmarshal(data []byte) error {
	c.PacketType = binary.BigEndian.Uint16(data[0:2])
	c.HashLength = binary.BigEndian.Uint16(data[2:4])
	hOffset := c.HashLength + 4
	iOffset := hOffset + 36 // 36 is a length of ID
	copy(c.Hash, data[4:hOffset])
	copy(c.ID, data[hOffset:iOffset])
	copy(c.Payload, data[iOffset:])
	return nil
}
