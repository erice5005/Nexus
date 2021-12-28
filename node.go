package nexus

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/ipv4"
)

type NodeConfig struct {
	Groups int
	Intf   *net.Interface
}
type Node struct {
	Id             string
	Config         NodeConfig
	Network        map[string]NetworkedNode
	BroadcastGroup ConnectedNexusGroup
	Groups         map[string]ConnectedNexusGroup
	ConnectStream  chan NexusGroup
	Locks          NodeLocks
}

type NodeLocks struct {
	NetworkMU sync.RWMutex
	GroupsMU  sync.RWMutex
}
type ConnectedNexusGroup struct {
	Group         NexusGroup
	Output        chan NodeMessage
	Input         chan interface{}
	InputHandler  func(out chan interface{}, cm *ipv4.ControlMessage, data []byte, src net.Addr)
	OutputHandler func(out chan interface{})
}

type NetworkedNode struct {
	Id       string
	LastSeen time.Time
	Groups   map[string]NexusGroup
}

type NexusGroup struct {
	Id         string
	Address    NexusGroupAddress
	Connection NexusGroupConnection
	Private    bool
}
type NexusGroupConnection struct {
	IPV4      *ipv4.PacketConn
	Conn      net.PacketConn
	Connected bool
}
type NexusGroupAddress struct {
	IP      net.IP
	Port    int
	Address *net.UDPAddr
}

type NodeMessage struct {
	SenderId string
	Routing  MessageRouting
	Data     map[string]interface{}
}

type NodeMessageType int64

const (
	Ping       NodeMessageType = 0
	Connect    NodeMessageType = 1
	Disconnect NodeMessageType = 2
	Message    NodeMessageType = 3
	GroupInfo  NodeMessageType = 4
)

type MessageRouting struct {
	DestinationAddress *net.UDPAddr
	DestinationId      string
	SourceAddress      *net.UDPAddr
	MessageType        NodeMessageType
}

func (msg NodeMessage) toJSON() []byte {
	marshed, err := json.Marshal(msg)
	if err != nil {
		return []byte("")
	}

	return marshed
}

func MsgFromJson(data []byte) NodeMessage {
	var msg NodeMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return NodeMessage{}
	}

	return msg
}

func NewNode(conf NodeConfig, id string) *Node {
	nodeid := id
	if nodeid == "" {
		nodeid = uuid.NewString()
	}
	var castIntf *net.Interface
	itfs, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for it := range itfs {
		addrs, _ := itfs[it].Addrs()
		if itfs[it].Flags&net.FlagMulticast != 0 && itfs[it].HardwareAddr != nil && addrs != nil {
			castIntf = &itfs[it]
		}
	}
	conf.Intf = castIntf

	n := &Node{
		Id:            nodeid,
		Config:        conf,
		Network:       make(map[string]NetworkedNode),
		Groups:        make(map[string]ConnectedNexusGroup),
		ConnectStream: make(chan NexusGroup, 10),
	}

	n.BroadcastGroup = ConnectedNexusGroup{
		Group: NewGroup("broadcast", NexusGroupAddress{
			IP:   net.IPv4(224, 0, 0, 250),
			Port: 9999,
		}),
		Output:       make(chan NodeMessage, 10000),
		Input:        make(chan interface{}, 10),
		InputHandler: n.BroadcastInputHandler,
	}

	return n
}

func (n *Node) HandleConnectStream() {
	for ng := range n.ConnectStream {
		if ng.Private {
			continue
		}
		n.ConnectGroup(ng)
	}
}

func NewGroup(id string, conf NexusGroupAddress) NexusGroup {
	groupid := id
	if groupid == "" {
		groupid = uuid.NewString()
	}
	if conf.Address == nil {
		conf.Address = &net.UDPAddr{
			IP:   conf.IP,
			Port: conf.Port,
		}
	}

	c, err := net.ListenPacket("udp", conf.Address.String())
	if err != nil {
		log.Printf("err: %v\n", err)
		return NexusGroup{}
	}

	p := ipv4.NewPacketConn(c)
	if conf.Port == 0 {
		hostPortVal := p.LocalAddr()
		hostSplit := strings.Split(hostPortVal.String(), ":")
		if len(hostSplit) == 2 {
			convd, _ := strconv.Atoi(hostSplit[1])
			conf.Port = convd
			conf.Address.Port = conf.Port
		}
	}

	return NexusGroup{
		Id:      groupid,
		Address: conf,
		Connection: NexusGroupConnection{
			Conn: c,
			IPV4: p,
		},
	}

}

func (ng NexusGroup) Connect(intf *net.Interface) {
	if err := ng.Connection.IPV4.JoinGroup(intf, ng.Address.Address); err != nil {
		log.Printf("failed to join group: %v with err: %v\n", ng.Address.Address, err)
		return
	}
	if err := ng.Connection.IPV4.SetControlMessage(ipv4.FlagDst, true); err != nil {
		log.Printf("failed to set control dst msg on group: %v with err: %v\n", ng.Address.Address, err)
	}
	if err := ng.Connection.IPV4.SetControlMessage(ipv4.FlagSrc, true); err != nil {
		log.Printf("failed to set control src msg on group: %v with err: %v\n", ng.Address.Address, err)
	}
	if err := ng.Connection.IPV4.SetMulticastInterface(intf); err != nil {
		log.Printf("write err: %v\n", err)
	}
}

// TODO: Leave group

func (ng ConnectedNexusGroup) Run(senderId string) {
	go func() {
		b := make([]byte, 1500)
		for {
			n, cm, src, err := ng.Group.Connection.IPV4.ReadFrom(b)
			if err != nil {
				continue
			}
			ng.InputHandler(ng.Input, cm, b[:n], src)
		}
	}()
	log.Printf("Started input\n")
	// go func() {
	for msg := range ng.Output {
		// log.Printf("msg: %v\n", msg)
		cm := &ipv4.ControlMessage{}
		msg.SenderId = senderId

		cm.Dst = ng.Group.Address.IP
		_, err := ng.Group.Connection.IPV4.WriteTo(msg.toJSON(), cm, ng.Group.Address.Address)
		if err != nil {
			continue
		}
	}
	// }()
	log.Printf("run ended: %v\n", ng.Group.Id)
}

func (n *Node) BroadcastInputHandler(inputStream chan interface{}, cm *ipv4.ControlMessage, data []byte, src net.Addr) {
	msg := MsgFromJson(data)
	if msg.SenderId == "" {
		return
	}
	if msg.SenderId == n.Id {
		return
	}

	var sourceNode NetworkedNode
	n.Locks.NetworkMU.RLock()
	if val, ok := n.Network[msg.SenderId]; ok {
		sourceNode = val
	} else {
		sourceNode = NetworkedNode{
			Id:     msg.SenderId,
			Groups: make(map[string]NexusGroup),
		}
	}
	n.Locks.NetworkMU.RUnlock()
	sourceNode.LastSeen = time.Now()
	switch msg.Routing.MessageType {
	case Ping:
		return
	case Connect:
		if len(n.Groups) > 0 {
			outputSet := make([]NexusGroup, 0)
			for _, ng := range n.Groups {
				if ng.Group.Private {
					continue
				}

				outputSet = append(outputSet, NexusGroup{
					Address: ng.Group.Address,
					Id:      ng.Group.Id,
					Private: ng.Group.Private,
				})
			}
			n.BroadcastGroup.Output <- NodeMessage{
				SenderId: n.Id,
				Routing: MessageRouting{
					MessageType: GroupInfo,
				},
				Data: map[string]interface{}{
					"groups": outputSet,
				},
			}
		}

		return
	case Message:
		inputStream <- msg.Data
	case GroupInfo:
		marshed, _ := json.Marshal(msg.Data["groups"])
		var groupList []NexusGroup
		json.Unmarshal(marshed, &groupList)
		for _, gi := range groupList {
			g := gi
			n.Locks.GroupsMU.Lock()
			if _, ok := n.Groups[g.Id]; !ok {
				n.Groups[g.Id] = ConnectedNexusGroup{
					Group: g,
				}
			}
			sourceNode.Groups[g.Id] = g
			n.Locks.GroupsMU.Unlock()
		}
	}
	n.Locks.NetworkMU.Lock()
	n.Network[msg.SenderId] = sourceNode
	n.Locks.NetworkMU.Unlock()

}

func (n *Node) Run() {
	go n.HandleBroadcastStream()
	go n.HandleConnectStream()
}

func (n *Node) HandleBroadcastStream() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		tryLimit := 3
		tryCount := 0

		for range ticker.C {
			if len(n.Groups) > 0 {
				break
			}

			tryCount++
			if tryCount == tryLimit {
				break
			}
		}

		if len(n.Groups) == 0 {
			ng := NewGroup("", NexusGroupAddress{
				IP:   net.IPv4(224, 0, 0, 1),
				Port: 0,
			})
			n.Groups[ng.Id] = ConnectedNexusGroup{
				Group: ng,
			}
			n.ConnectStream <- ng
		} else {
			for _, ng := range n.Groups {
				if !ng.Group.Connection.Connected && !ng.Group.Private {
					n.ConnectStream <- ng.Group
				}
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			n.BroadcastGroup.Output <- NodeMessage{
				SenderId: n.Id,
				Routing: MessageRouting{
					DestinationAddress: n.BroadcastGroup.Group.Address.Address,
					MessageType:        Ping,
				},
			}
		}
	}()
	n.BroadcastGroup.InputHandler = n.BroadcastInputHandler
	n.BroadcastGroup.Group.Connect(n.Config.Intf)
	n.BroadcastGroup.Group.Connection.Connected = true
	n.BroadcastGroup.Output <- NodeMessage{
		SenderId: n.Id,
		Routing: MessageRouting{
			DestinationAddress: n.BroadcastGroup.Group.Address.Address,
			MessageType:        Connect,
		},
	}
	n.BroadcastGroup.Run(n.Id)
}

func (n *Node) ConnectGroup(ngc NexusGroup) {
	ng := ConnectedNexusGroup{
		Group:  ngc,
		Output: make(chan NodeMessage, 10000),
		Input:  make(chan interface{}, 10),
	}
	log.Printf("N: %v, NGC: %v\n", n.Id, ngc.Id)
	ng.Group.Connect(n.Config.Intf)
	ng.Group.Connection.Connected = true
	if ng.InputHandler == nil {
		ng.InputHandler = n.BroadcastInputHandler
	}
	go func() {
		ng.Run(n.Id)
	}()
	ng.Output <- NodeMessage{
		SenderId: n.Id,
		Routing: MessageRouting{
			DestinationAddress: ng.Group.Address.Address,
			MessageType:        Connect,
		},
	}
	n.Groups[ng.Group.Id] = ng
	log.Print("Group connected\n")
}
