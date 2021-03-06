package ptp

import (
	"net"
	"reflect"
	"sync"
	"testing"
)

func TestGenerateMac(t *testing.T) {
	macs := make(map[string]net.HardwareAddr)

	for i := 0; i < 10000; i++ {
		smac, mac := GenerateMAC()
		if smac == "" {
			t.Errorf("Failed to generate mac")
			return
		}
		_, e := macs[smac]
		if e {
			t.Errorf("Same MAC was generated")
			return
		}
		macs[smac] = mac
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"t1", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateToken(); got == tt.want {
				t.Errorf("GenerateToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPrivateIP(t *testing.T) {
	type args struct {
		ip net.IP
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"Empty IP", args{}, false, true},
		{"10.x subnet", args{net.ParseIP("10.12.13.14")}, true, false},
		{"10.x subnet", args{net.ParseIP("10.0.0.1")}, true, false},
		{"172.16.x subnet", args{net.ParseIP("172.16.17.18")}, true, false},
		{"172.16.x subnet", args{net.ParseIP("172.16.0.1")}, true, false},
		{"192.168.x subnet", args{net.ParseIP("192.168.0.1")}, true, false},
		{"192.168.x subnet", args{net.ParseIP("192.168.1.1")}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isPrivateIP(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("isPrivateIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isPrivateIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringifyState(t *testing.T) {
	type args struct {
		state PeerState
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Stringify state: Init", args{PeerStateInit}, "INITIALIZING"},
		{"Stringify state: Waiting IP", args{PeerStateRequestedIP}, "WAITING_IP"},
		{"Stringify state: Requesting Proxies", args{PeerStateRequestingProxy}, "REQUESTING_PROXIES"},
		{"Stringify state: Waiting Proxies", args{PeerStateWaitingForProxy}, "WAITING_PROXIES"},
		{"Stringify state: Waiting Connection", args{PeerStateWaitingToConnect}, "WAITING_CONNECTION"},
		{"Stringify state: Initializing Connection", args{PeerStateConnecting}, "INITIALIZING_CONNECTION"},
		{"Stringify state: Connected", args{PeerStateConnected}, "CONNECTED"},
		{"Stringify state: Disconnected", args{PeerStateDisconnect}, "DISCONNECTED"},
		{"Stringify state: Stopped", args{PeerStateStop}, "STOPPED"},
		{"Stringify state: Cooldown", args{PeerStateCooldown}, "COOLDOWN"},
		{"Stringify state: Unknown", args{PeerState(99)}, "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringifyState(tt.args.state); got != tt.want {
				t.Errorf("StringifyState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsInterfaceLocal(t *testing.T) {
	type args struct {
		ip net.IP
	}
	ActiveInterfaces = []net.IP{net.ParseIP("10.10.10.1")}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Interface in list", args{net.ParseIP("10.10.10.1")}, true},
		{"Interface not in list", args{net.ParseIP("192.168.0.1")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInterfaceLocal(tt.args.ip); got != tt.want {
				t.Errorf("IsInterfaceLocal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_min(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"min 1", args{0, 0}, 0},
		{"min 2", args{0, 1}, 0},
		{"min 3", args{1, 0}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := min(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSrvLookup(t *testing.T) {
	type args struct {
		name   string
		proto  string
		domain string
	}
	res := make(map[int]string)
	res[0] = "prod-bazaar-eu-1.s.optdyn.com.:6881"
	tests := []struct {
		name    string
		args    args
		want    map[int]string
		wantErr bool
	}{
		{"Wrong name", args{"boogie", "tcp", "subutai.io"}, nil, true},
		{"Wrong protocol", args{"dht", "boogie", "subutai.io"}, nil, true},
		{"Wrong domain", args{"dht", "tcp", "subutai.subutai"}, nil, true},
		{"Positive result", args{"dht", "tcp", "subutai.io"}, res, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SrvLookup(tt.args.name, tt.args.proto, tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("SrvLookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SrvLookup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerToPeer_FindNetworkAddresses(t *testing.T) {
	type fields struct {
		UDPSocket       *Network
		LocalIPs        []net.IP
		Dht             *DHTClient
		Crypter         Crypto
		Shutdown        bool
		ForwardMode     bool
		ReadyToStop     bool
		MessageHandlers map[uint16]MessageHandler
		PacketHandlers  map[PacketType]PacketHandlerCallback
		PeersLock       sync.Mutex
		Hash            string
		Interface       TAP
		Peers           *Swarm
		HolePunching    sync.Mutex
		ProxyManager    *ProxyManager
		outboundIP      net.IP
		UsePMTU         bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"Passing", fields{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PeerToPeer{
				UDPSocket:       tt.fields.UDPSocket,
				LocalIPs:        tt.fields.LocalIPs,
				Dht:             tt.fields.Dht,
				Crypter:         tt.fields.Crypter,
				Shutdown:        tt.fields.Shutdown,
				ForwardMode:     tt.fields.ForwardMode,
				ReadyToStop:     tt.fields.ReadyToStop,
				MessageHandlers: tt.fields.MessageHandlers,
				PacketHandlers:  tt.fields.PacketHandlers,
				Hash:            tt.fields.Hash,
				Interface:       tt.fields.Interface,
				Swarm:           tt.fields.Peers,
				HolePunching:    tt.fields.HolePunching,
				ProxyManager:    tt.fields.ProxyManager,
				outboundIP:      tt.fields.outboundIP,
				UsePMTU:         tt.fields.UsePMTU,
			}
			if err := p.FindNetworkAddresses(); (err != nil) != tt.wantErr {
				t.Errorf("PeerToPeer.FindNetworkAddresses() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_isDeviceExists(t *testing.T) {
	type args struct {
		name string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty name", args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDeviceExists(tt.args.name); got != tt.want {
				t.Errorf("isDeviceExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIntroString(t *testing.T) {
	type args struct {
		intro string
	}

	hs0 := new(PeerHandshake)
	hs0.ID = "1"
	hs0.IP = net.ParseIP("10.11.12.13")
	hs0.HardwareAddr, _ = net.ParseMAC("00:11:22:33:44:55")
	hs0.Endpoint, _ = net.ResolveUDPAddr("udp4", "192.168.0.1:1234")

	tests := []struct {
		name    string
		args    args
		want    *PeerHandshake
		wantErr bool
	}{
		{"empty test", args{}, nil, true},
		{"broken message", args{",,"}, nil, true},
		{"broken mac", args{",001122334455,,"}, nil, true},
		{"broken ip", args{",00:11:22:33:44:55,a,"}, nil, true},
		{"broken udp addr", args{",00:11:22:33:44:55,10.11.12.13,a:b"}, nil, true},
		{"passing", args{"1,00:11:22:33:44:55,10.11.12.13,192.168.0.1:1234"}, hs0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIntroString(tt.args.intro)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIntroString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseIntroString() = %v, want %v", got, tt.want)
			}
		})
	}
}
