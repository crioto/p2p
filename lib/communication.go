package ptp

// Communication subsystem
type Communication struct {
}

// handle will receipe a message, unmarshal it and decide
// what to do next with it
func (c *Communication) handle(msg *Message) error {
	comm := new(CommunicationPacket)
	err := comm.Unmarshal(msg.Payload)
	if err != nil {
		Log(Error, "Failed to unmarshal comm: %s", err.Error())
		return err
	}

	// Thre is a lot of different packet types
	// TODO: Wrap it as constants
	switch comm.PacketType {
	case 0: // Ping message
		break
	}

	return nil
}
