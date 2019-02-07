package ptp

import (
	"fmt"
	"net"
)

type Net struct {
	port    uint16
	rPort   uint16
	socket  *net.UDPConn
	running bool
	data    chan *Message
	comm    chan *Message
}

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

func (n *Net) Close() error {
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

func (n *Net) listen(receivedCallback UDPReceivedCallback) error {
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
