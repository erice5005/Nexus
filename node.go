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

type NodeMsg struct {
	Dst *net.UDPAddr
	Src *net.UDPAddr
	Msg map[string]string
}

func (nm NodeMsg) ToJson() string {
	encd, err := json.Marshal(nm)
	if err != nil {
		return ""
	}
	return string(encd)
}

func MsgFromJson(inp []byte) NodeMsg {
	var out NodeMsg
	err := json.Unmarshal(inp, &out)
	if err != nil {
		return NodeMsg{}
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
	ReadStream    chan NodeMsg
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
	Id       string
	LastSeen time.Time
}

func (mg MulticastGroup) Stream(readStream chan NodeMsg, nodeId string) {
	log.Printf("start stream: %v\n", mg.Id)
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
			if msg.Msg["id"] == nodeId {
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
					// log.Printf("Group: %v\n", cm)

					// log.Printf("Group %v got MSG %v\n", msg.Msg["groupid"], msg)
					readStream <- msg
					// joined group, do something
				} else {
					log.Printf("unknown msg: %v\n", string(b[:n]))
					// unknown group, discard
					continue
				}
			}
		}
	}()
	go func() {
		testTicker := time.NewTicker(1 * time.Second)
		for range testTicker.C {
			mg.WriteStream <- []byte(
				NodeMsg{
					Msg: map[string]string{
						"id":      nodeId,
						"groupid": mg.Id,
						"type":    "ping",
						// "string": "i'm a test message",
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
			n.Broadcast.WriteStream <- []byte(
				NodeMsg{
					Msg: map[string]string{
						"id": n.Id,
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
		if msg.Msg["id"] == n.Id {
			continue
		}

		// log.Printf("Got: %v\n", msg)
		if _, ok := n.Network[msg.Msg["id"]]; !ok {
			log.Printf("I'm new: %v\n", msg.Msg["id"])
			n.Network[msg.Msg["id"]] = NetworkedNode{
				Id: msg.Msg["id"],
			}
			// log.Printf("group len: %v\n", len(n.Groups))
			if len(n.Groups) > 0 {
				for _, ngg := range n.Groups {
					// toSend, _ := json.Marshal(ngg)
					// log.Printf("ngg: %v\n", ngg)
					n.Broadcast.WriteStream <- []byte(NodeMsg{
						Msg: map[string]string{
							"id":          n.Id,
							"groupip":     ngg.IP.String(),
							"groupport":   strconv.Itoa(ngg.Port),
							"groupid":     ngg.Id,
							"isGroupInfo": "true",
						},
					}.ToJson())
				}
			}
		}
		// mark off the last time the node has been seen
		nn := n.Network[msg.Msg["id"]]
		nn.LastSeen = time.Now()
		n.Network[msg.Msg["id"]] = nn

		// check for group info in the message
		// it it exists, create a new group and add it to the list
		// log.Printf("MSG: %v\n", msg)
		if msg.Dst != nil {
			continue
		}
		if val, ok := msg.Msg["isGroupInfo"]; ok {
			if val != "true" {
				continue
			}

			groupIP := net.ParseIP(msg.Msg["groupip"])
			groupPortString := msg.Msg["groupport"]
			groupID := msg.Msg["groupid"]

			groupPort, _ := strconv.Atoi(groupPortString)

			if _, ok := n.Groups[groupID]; !ok {
				if len(n.Groups) <= n.GroupLimit {
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
	}
}

func (bs *BroadcastStream) Listen(st chan NodeMsg) {
	for {
		b := make([]byte, 2500)
		n, src, err := bs.ListenConn.ReadFromUDP(b) //n, src
		if err != nil {
			log.Printf("Broadcast read err: %v\n", err)
		}
		msg := MsgFromJson(b[:n])
		msg.Src = src

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
		ReadStream:    make(chan NodeMsg, 10),
		Network:       make(map[string]NetworkedNode),
		Groups:        make(map[string]MulticastGroup),
		ConnectStream: make(chan string, 10),
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
