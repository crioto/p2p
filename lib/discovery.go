package ptp

import (
	"fmt"
	"net"
	"sync"
)

// Discovery service used as a randevouz service
// It used for peer discovery
// When two peers want to connect to each other, Discovery service
// will take request from one peer and sent it to another peer
// If second peer replies, then first peer sends encrypted information
// about it endpoints to the second peer and second peer does the same
// upon receival.
// Discovery service operates with peers as with nodes. For each node
// it knows only it's ID (assigned by Discovery service) and the
// endpoint
type Discovery struct {
	nodes  map[string]*net.UDPAddr
	lock   sync.RWMutex
	socket *Net
}

// Init will allocate map for nodes and establish a connection with other
// discovery service nodes. The first nodes it connects to are being global
// nodes, discovered with
func (d *Discovery) Init(socket *Net) error {
	if socket == nil {
		return fmt.Errorf("nil socket")
	}
	d.socket = socket
	d.nodes = make(map[string]*net.UDPAddr)

	// TODO: Do srv lookup there

	return nil
}

// Close will stop discovery service
func (d *Discovery) Close() error {
	return nil
}

// find does a lookup of a peer. If peer found it send
func (d *Discovery) find(id string) error {

	return nil
}
