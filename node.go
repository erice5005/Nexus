package nexus

import (
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/ipv4"
)

// type NodeMessage struct {
// 	Dst *net.UDPAddr
// 	Src *net.UDPAddr
// 	Msg map[string]string
// }
type NodeMessage struct {
	OriginId        string
	DestinationAddr *net.UDPAddr
	DestinationId   string
	SourceAddr      *net.UDPAddr
	Msg             map[string]string
}

func (nm NodeMessage) ToJson() string {
	encd, err := json.Marshal(nm)
	if err != nil {
		return ""
	}
	return string(encd)
}

func MsgFromJson(inp []byte) NodeMessage {
	var out NodeMessage
	err := json.Unmarshal(inp, &out)
	if err != nil {
		return NodeMessage{}
	}
	return out
}

type Node struct {
	Id string
	// Broadcast     *BroadcastStream
	Broadcast     *MulticastGroup
	Groups        map[string]MulticastGroup
	Intf          *net.Interface
	Network       map[string]NetworkedNode
	ReadStream    chan NodeMessage
	ConnectStream chan string
	GroupLimit    int
}

type BroadcastStream struct {
	Addr        *net.UDPAddr
	WriteConn   *net.UDPConn
	ListenConn  *net.UDPConn
	WriteStream chan []byte
}

type MulticastGroup struct {
	IP          net.IP
	Port        int
	Conn        net.PacketConn
	IPV4        *ipv4.PacketConn
	Connected   bool
	Addr        *net.UDPAddr
	Id          string
	WriteStream chan []byte
}

type NetworkedNode struct {
	Id        string
	LastSeen  time.Time
	SeenNodes map[string]NetworkedNode
}

func (mg MulticastGroup) Stream(readStream chan NodeMessage, node *Node) {
	go func() {
		b := make([]byte, 1500)
		for {
			n, cm, src, err := mg.IPV4.ReadFrom(b)
			if err != nil {
				log.Printf("read err: %v\n", err)
				continue
				// error handling
			}
			msg := MsgFromJson(b[:n])
			if msg.OriginId == "" {
				continue
			}
			if msg.OriginId == node.Id {
				continue
			}
			if src == mg.IPV4.LocalAddr() {
				continue
			}
			// log.Printf("dst: %v\n", cm)
			if cm.Dst == nil {
				continue
			}
			if cm.Dst.IsMulticast() {
				if cm.Dst.Equal(mg.Addr.IP) {
					readStream <- msg
				} else {
					log.Printf("unknown msg: %v\n", string(b[:n]))
					continue
				}
			}
		}
	}()
	go func() {
		testTicker := time.NewTicker(10 * time.Second)
		for range testTicker.C {
			marshedNet, _ := json.Marshal(node.Network)
			mg.WriteStream <- []byte(
				NodeMessage{
					OriginId: node.Id,
					Msg: map[string]string{
						"id":        node.Id,
						"groupid":   mg.Id,
						"type":      "ping",
						"seenNodes": string(marshedNet),
					},
				}.ToJson(),
			)
		}
	}()
	func() {
		for msg := range mg.WriteStream {
			// log.Printf("writing msg: %v\n", string(msg))
			_, err := mg.IPV4.WriteTo(msg, nil, mg.Addr)
			if err != nil {
				log.Printf("write err: %v\n", err)
			}
		}
	}()
	log.Printf("ended: %v\n", time.Now())
}

func NewBroadcastStream(addr string) *BroadcastStream {
	targetAddr, _ := net.ResolveUDPAddr("udp", addr)

	writeConn, _ := net.DialUDP("udp", nil, targetAddr)
	listenConn, _ := net.ListenMulticastUDP("udp", nil, targetAddr)
	listenConn.SetReadBuffer(2500)

	return &BroadcastStream{
		Addr:        targetAddr,
		WriteConn:   writeConn,
		ListenConn:  listenConn,
		WriteStream: make(chan []byte, 10),
	}

}

func (n *Node) RunBroadcastNet() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {

			marshedNet, _ := json.Marshal(n.Network)

			n.Broadcast.WriteStream <- []byte(
				NodeMessage{
					OriginId: n.Id,
					Msg: map[string]string{
						"id":        n.Id,
						"type":      "seenUpdate",
						"seenNodes": string(marshedNet),
					},
				}.ToJson(),
			)
		}
	}()
	// go n.Broadcast.Listen(n.ReadStream)
	// go n.Broadcast.Write()
}

func (n *Node) GroupInitWatch() {
	ticker := time.NewTicker(1 * time.Second)
	tryLimit := 3
	tryCount := 0
	for range ticker.C {
		if len(n.Groups) > 0 {
			break
		}
		// if len(n.Network) == 0 {
		tryCount++
		log.Printf("Seen: %v nets and %v groups\n", len(n.Network), len(n.Groups))
		// Other nodes exist, pull some data from them
		// }
		if tryCount == tryLimit {
			break
			// Hasn't found anything, need a group created
		}
	}
	log.Printf("Seen: %v groups\n", len(n.Groups))
	if len(n.Groups) == 0 {
		log.Printf("Needs a fresh group\n")
		nid := n.NewGroup("", net.IPv4(224, 0, 0, byte(rand.Intn(249))), 0)
		log.Printf("NID: %v\n", nid)
		n.ConnectToGroup(nid)
	} else {
		for _, gg := range n.Groups {
			log.Printf("connect to: %v\n", gg)
			if !gg.Connected {
				n.ConnectToGroup(gg.Id)
			}
		}
	}
}

func (n *Node) ReadStreamHandler() {
	for msg := range n.ReadStream {
		if msg.OriginId == n.Id {
			continue
		}

		// log.Printf("Got: %v\n", msg)
		if _, ok := n.Network[msg.OriginId]; !ok {
			log.Printf("I'm new: %v\n", msg.OriginId)
			n.Network[msg.OriginId] = NetworkedNode{
				Id:        msg.OriginId,
				SeenNodes: make(map[string]NetworkedNode),
			}
			// if len(n.Groups) > 0 {
			for _, ngg := range n.Groups {
				n.Broadcast.WriteStream <- []byte(NodeMessage{
					OriginId: n.Id,
					Msg: map[string]string{
						"id":          n.Id,
						"groupip":     ngg.IP.String(),
						"groupport":   strconv.Itoa(ngg.Port),
						"groupid":     ngg.Id,
						"isGroupInfo": "true",
					},
				}.ToJson())
			}
			// }
		}
		// mark off the last time the node has been seen
		nn := n.Network[msg.OriginId]
		nn.LastSeen = time.Now()
		n.Network[msg.OriginId] = nn

		// check for group info in the message
		// it it exists, create a new group and add it to the list
		// log.Printf("MSG: %v\n", msg)
		if msg.DestinationAddr != nil {
			continue
		}
		if val, ok := msg.Msg["isGroupInfo"]; ok {
			if val != "true" {
				continue
			}
			log.Printf("Got group info: %v\n", msg.Msg)
			groupIP := net.ParseIP(msg.Msg["groupip"])
			groupPortString := msg.Msg["groupport"]
			groupID := msg.Msg["groupid"]

			groupPort, _ := strconv.Atoi(groupPortString)

			if _, ok := n.Groups[groupID]; !ok {
				if len(n.Groups) <= n.GroupLimit || n.GroupLimit == 0 {
					log.Printf("adding group: %v\n", groupID)
					n.NewGroup(groupID, groupIP, groupPort)
				} else {
					log.Printf("Needs a fresh group\n")
					nid := n.NewGroup("", net.IPv4(224, 0, 0, byte(rand.Intn(249))), 0)
					log.Printf("NID: %v\n", nid)
					n.ConnectToGroup(nid)
				}

			}
		}
		if val, ok := msg.Msg["seenNodes"]; ok {
			log.Printf("val: %v\n", string(val))
			var seenNodes map[string]NetworkedNode
			err := json.Unmarshal([]byte(val), &seenNodes)
			if err != nil {
				log.Printf("error unmarshalling seennodes from: %v\n", val)
				continue
			}
			nn := n.Network[msg.OriginId]
			nn.SeenNodes = seenNodes
			n.Network[msg.OriginId] = nn
		}
	}
}

func (bs *BroadcastStream) Listen(st chan NodeMessage) {
	for {
		b := make([]byte, 2500)
		n, src, err := bs.ListenConn.ReadFromUDP(b) //n, src
		if err != nil {
			log.Printf("Broadcast read err: %v\n", err)
		}
		msg := MsgFromJson(b[:n])
		msg.SourceAddr = src

		st <- msg
		// log.Printf("%v bytes from %v: %v\n", n, src, string(b[:n]))
	}
}

func (bs *BroadcastStream) Write() {
	for msg := range bs.WriteStream {
		bs.WriteConn.Write(msg)
	}
}

func NewNode(broadcastAddr string) *Node {
	nodeid := uuid.NewString()
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

	// broadcastNet := NewBroadcastStream(broadcastAddr)
	return &Node{
		Id:   nodeid,
		Intf: castIntf,
		// Broadcast:     broadcastNet,
		ReadStream:    make(chan NodeMessage, 10),
		Network:       make(map[string]NetworkedNode),
		Groups:        make(map[string]MulticastGroup),
		ConnectStream: make(chan string, 10),
		GroupLimit:    0,
	}
}

func (n *Node) NewGroup(id string, targetIP net.IP, port int) string {

	// n.Groups[id] = MulticastGroup{
	// 	IP:          targetIP,
	// 	Port:        port,
	// 	Conn:        c,
	// 	IPV4:        p,
	// 	Addr:        tg,
	// 	Id:          id,
	// 	WriteStream: make(chan []byte, 10),
	// }
	if id == "" {
		id = uuid.NewString()
	}
	n.Groups[id] = NewGroup(id, targetIP, port)
	return id
}

func NewGroup(id string, targetIP net.IP, port int) MulticastGroup {
	c, err := net.ListenPacket("udp", net.JoinHostPort("224.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return MulticastGroup{}
	}

	p := ipv4.NewPacketConn(c)

	if port == 0 {
		hostPortVal := p.LocalAddr()
		hostSplit := strings.Split(hostPortVal.String(), ":")
		if len(hostSplit) == 2 {
			convd, _ := strconv.Atoi(hostSplit[1])
			port = convd
		}
	}

	tg := &net.UDPAddr{
		IP:   targetIP,
		Port: port,
	}

	return MulticastGroup{
		IP:          targetIP,
		Port:        port,
		Conn:        c,
		IPV4:        p,
		Addr:        tg,
		Id:          id,
		WriteStream: make(chan []byte, 10),
	}
}

func (n *Node) ConnectToGroup(id string) {
	targetGroup := n.Groups[id]

	// tg := n.Groups[id].Addr
	ConnectToGroup(targetGroup, n.Intf)
	log.Printf("finished control flags\n")
	targetGroup.Connected = true
	// targetGroup.Stream(n.ReadStream)
	n.Groups[id] = targetGroup
	n.ConnectStream <- id
	log.Printf("group connected: %v\n", id)
}

func ConnectToGroup(tg MulticastGroup, intf *net.Interface) {
	log.Printf("target group: %v\n", tg)
	if err := tg.IPV4.JoinGroup(intf, tg.Addr); err != nil {
		log.Printf("failed to join group: %v with err: %v\n", tg.Addr, err)
		return
	}

	if err := tg.IPV4.SetControlMessage(ipv4.FlagDst, true); err != nil {
		log.Printf("failed to set control dst msg on group: %v with err: %v\n", tg.Addr, err)
	}
	if err := tg.IPV4.SetControlMessage(ipv4.FlagSrc, true); err != nil {
		log.Printf("failed to set control src msg on group: %v with err: %v\n", tg.Addr, err)
	}
	if err := tg.IPV4.SetMulticastInterface(intf); err != nil {
		// error handling
		log.Printf("write err: %v\n", err)
	}
	// if err := tg.IPV4.SetMulticastLoopback(false); err != nil {
	// 	log.Printf("write err: %v\n", err)
	// }
}
