package ptp

import (
	"fmt"
	"net"
)

// Net is a networking subsystem
type Net struct {
	port    uint16        // listening port
	rPort   uint16        // remote port reported by discovery service
	socket  *net.UDPConn  // network socket
	running bool          // whether listener is running or not
	data    chan *Message // channel to pass LAN messages to the system
	comm    chan *Message // channel to pass communication message to the subsystem
}

// Init will initialize networking subsystem
func (n *Net) Init(port uint16) error {
	n.running = false
	n.port = port
	n.data = make(chan *Message)
	n.comm = make(chan *Message)

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	n.socket, err = net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}

	return nil
}

// Close will stop listener, data and comm channels
func (n *Net) Close() error {
	if !n.running {
		return fmt.Errorf("already closed")
	}
	n.running = false
	if n.socket != nil {
		err := n.socket.Close()
		if err != nil {
			Log(Error, "Failed to close socket: %s", err.Error())
		}
		n.socket = nil
	}
	if n.data != nil {
		close(n.data)
		n.data = nil
	}
	if n.comm != nil {
		close(n.comm)
		n.comm = nil
	}
	return nil
}

func (n *Net) listen() error {
	Log(Info, "Starting UDP listener")
	if n.socket == nil {
		return fmt.Errorf("nil socket")
	}
	n.running = true
	var buffer [2048]byte
	for n.running {
		_, src, err := n.socket.ReadFromUDP(buffer[:])
		if err != nil {
			Log(Warning, "Failed to read from socket: %s", err.Error())
			continue
		}
		msg := new(Message)
		err = msg.Unmarshal(buffer[:])
		if err != nil {
			Log(Warning, "unmarshal failed: %s", err.Error())
			continue
		}
		n.route(src, msg)
	}
	Log(Info, "Stopping UDP Listener")
	return nil
}

func (n *Net) route(src *net.UDPAddr, msg *Message) {
	if msg.MagicCookie == MagicCookieLan {
		n.data <- msg
		return
	}
	if msg.MagicCookie == MagicCookieComm {
		n.comm <- msg
		return
	}
}

// Send will marshal Message and send it to the destination
func (n *Net) Send(msg *Message, dst *net.UDPAddr) error {
	if msg == nil {
		return fmt.Errorf("nil message")
	}
	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("marshal failed")
	}
	return n.SendRaw(data, dst)
}

// SendRaw will send bytes over socket
func (n *Net) SendRaw(data []byte, dst *net.UDPAddr) error {
	if n.socket == nil {
		return fmt.Errorf("nil socket")
	}
	if data == nil {
		return fmt.Errorf("nil data")
	}
	if dst == nil {
		return fmt.Errorf("nil destination")
	}
	_, err := n.socket.WriteToUDP(data, dst)
	if err != nil {
		return err
	}
	return nil
}

// GetPorts returns local and remote ports
func (n *Net) GetPorts() (uint16, uint16) {
	return n.port, n.rPort
}
