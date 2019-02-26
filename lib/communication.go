package ptp

// Communication subsystem
type Communication struct {
}

// Init will initialize communication subsystem
func (c *Communication) Init() error {
	return nil
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
	case 0: // Ping Message
		break
	case 1: // Status report
		break
	case 2: // Latency
		break
	case 3: // Version mismatch
		break
	case 10: // Subnet request/response
		break
	case 11: // Request information about IP
		break
	case 12: // Notify about IP change/assign
		break
	case 13: // Notify about IP conflict
		break
	case 20: // Discovery initialization
		break
	case 21: // Find request
		break
	case 22: // Node request
		break
	case 23: // Proxy message
		break
	case 24: // IP request/response
		break
	}

	return nil
}
