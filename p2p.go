package main

import (
	"flag"
	"fmt"
	"github.com/danderson/tuntap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"p2p/dht"
	//	"strings"
	"p2p/udpcs"
	"time"
)

// Main structure
type PTPCloud struct {

	// IP Address assigned to device at startup
	IP string

	// MAC Address assigned to device or generated by the application (TODO: Implement random generation and MAC assignment)
	Mac string

	// Netmask for device
	Mask string

	// Name of the device
	DeviceName string

	// Path to tool that is used to configure network device (only "ip" tools is supported at this moment)
	IPTool string `yaml:"iptool"`

	// TUN/TAP Interface
	Interface *os.File

	// Representation of TUN/TAP Device
	Device *tuntap.Interface

	NetworkPeers []NetworkPeer
}

type NetworkPeer struct {
	PeerIP      net.IP
	PeerPort    int
	PeerLocalIP net.IP
	PeerHW      net.HardwareAddr
}

// Creates TUN/TAP Interface and configures it with provided IP tool
func (ptp *PTPCloud) CreateDevice(ip, mac, mask, device string) error {
	var err error

	ptp.IP = ip
	ptp.Mac = mac
	ptp.Mask = mask
	ptp.DeviceName = device

	// Extract necessary information from config file
	// TODO: Remove hard-coded path
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("[ERROR] Failed to load config: %v", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, ptp)
	if err != nil {
		log.Printf("[ERROR] Failed to parse config: %v", err)
		return err
	}

	ptp.Device, err = tuntap.Open(ptp.DeviceName, tuntap.DevTap)
	if ptp.Device == nil {
		log.Fatalf("[FATAL] Failed to open TAP device: %v", err)
		return err
	} else {
		log.Printf("[INFO] %v TAP Device created", ptp.DeviceName)
	}

	linkup := exec.Command(ptp.IPTool, "link", "set", "dev", ptp.DeviceName, "up")
	err = linkup.Run()
	if err != nil {
		log.Fatalf("[ERROR] Failed to up link: %v", err)
		return err
	}

	// Configure new device
	log.Printf("[INFO] Setting %s IP on device %s\n", ptp.IP, ptp.DeviceName)
	setip := exec.Command(ptp.IPTool, "addr", "add", ptp.IP+"/24", "dev", ptp.DeviceName)
	err = setip.Run()
	if err != nil {
		log.Fatalf("[FATAL] Failed to set IP: %v", err)
		return err
	}
	return nil
}

// Handles a packet that was received by TUN/TAP device
// Receiving a packet by device means that some application sent a network
// packet within a subnet in which our application works.
// This method calls appropriate gorouting for extracted packet protocol
func handlePacket(ptp *PTPCloud, contents []byte, proto int) {
	/*
		512   (PUP)
		2048  (IP)
		2054  (ARP)
		32821 (RARP)
		33024 (802.1q)
		34525 (IPv6)
		34915 (PPPOE discovery)
		34916 (PPPOE session)
	*/
	switch proto {
	case 512:
		log.Printf("[DEBUG] Received PARC Universal Packet")
	case 2048:
		ptp.handlePacketIPv4(contents)
	case 2054:
		log.Printf("[DEBUG] Received ARP Packet")
		ptp.handlePacketARP(contents)
	case 32821:
		log.Printf("[DEBUG] Received RARP Packet")
	case 33024:
		log.Printf("[DEBUG] Received 802.1q Packet")
	case 34525:
		ptp.handlePacketIPv6(contents)
	case 34915:
		log.Printf("[DEBUG] Received PPPoE Discovery Packet")
	case 34916:
		log.Printf("[DEBUG] Received PPPoE Session Packet")
	default:
		log.Printf("[DEBUG] Received Undefined Packet")
	}
}

// Listen TAP interface for incoming packets
func (ptp *PTPCloud) ListenInterface() {
	// Read packets received by TUN/TAP device and send them to a handlePacket goroutine
	// This goroutine will decide what to do with this packet
	for {
		packet, err := ptp.Device.ReadPacket()
		if err != nil {
			log.Printf("Error reading packet: %s", err)
		}
		if packet.Truncated {
			log.Printf("[DEBUG] Truncated packet")
		}
		// TODO: Make handlePacket as a part of PTPCloud
		go handlePacket(ptp, packet.Packet, packet.Protocol)
	}
}

func main() {
	// TODO: Move this to init() function
	var (
		argIp     string
		argMask   string
		argMac    string
		argDev    string
		argDirect string
		argHash   string
	)

	// TODO: Improve this
	flag.StringVar(&argIp, "ip", "none", "IP Address to be used")
	// TODO: Parse this properly
	flag.StringVar(&argMask, "mask", "none", "Network mask")
	// TODO: Implement this
	flag.StringVar(&argMac, "mac", "none", "MAC Address for a TUN/TAP interface")
	flag.StringVar(&argDev, "dev", "none", "TUN/TAP interface name")
	// TODO: Direct connection is not implemented yet
	flag.StringVar(&argDirect, "direct", "none", "IP to connect to directly")
	flag.StringVar(&argHash, "hash", "none", "Infohash for environment")

	flag.Parse()
	if argIp == "none" || argMask == "none" || argDev == "none" {
		fmt.Println("USAGE: p2p [OPTIONS]")
		fmt.Printf("\nOPTIONS:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	udp := new(udpcs.UDPClient)
	udp.Init("", 0)
	port := udp.GetPort()
	log.Printf("[INFO] Started UDP Listener at port %d", port)
	go udp.Listen(udpcs.Process_p2p_msg)

	// Create new DHT Client, configured it and initialize
	// During initialization procedure, DHT Client will send
	// a introduction packet along with a hash to a DHT bootstrap
	// nodes that was hardcoded into it's code
	var dhtClient dht.DHTClient
	config := dhtClient.DHTClientConfig()
	config.NetworkHash = argHash
	config.P2PPort = port
	dhtClient.Initialize(config)

	ptp := new(PTPCloud)
	ptp.CreateDevice(argIp, argMac, argMask, argDev)

	// Capture SIGINT
	// This is used for development purposes only, but later we should consider updating
	// this code to handle signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			fmt.Println("Received signal: ", sig)
			os.Exit(0)
		}
	}()

	go ptp.ListenInterface()
	for {
	}
}

func (ptp *PTPCloud) UpdatePeersTable(dht *dht.DHTClient) {
	//timestamp := time.Now().Local()
	for _ = range time.Tick(15 * time.Second) {
		if len(ptp.NetworkPeers) != len(dht.NetworkPeers) {
			//for _, peer := range dht.NetworkPeers {
			//			p := strings.Split(peer, ":")
			/*for _, lp := range ptp.NetworkPeers {
			}
			*/
			//}
		}
	}
}

// WriteToDevice writes data to created TUN/TAP device
func (ptp *PTPCloud) WriteToDevice(b []byte) {
	var p tuntap.Packet
	p.Protocol = 2054
	p.Truncated = false
	p.Packet = b
	if ptp.Device == nil {
		log.Printf("[ERROR] TUN/TAP Device not initialized")
		return
	}
	err := ptp.Device.WritePacket(&p)
	if err != nil {
		log.Printf("[ERROR] Failed to write to TUN/TAP device")
	}
}
