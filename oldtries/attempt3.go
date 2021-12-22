package main

// import (
// 	"net"
// 	"log"
// 	"strings"
// 	"strconv"
// 	"github.com/google/uuid"
// 	"golang.org/x/net/ipv4"
// 	"encoding/json"
// )

//Nexus Protocol
// - Node pulls available interfaces to connect over
// - Finds one supporting multicast, sets it as the interface for that node
// - Returns the node with generated ids, the interface, and additional metadata/config
// - All nodes need to be configured with the same starting broadcast network
// - On start node should connect to the broadcast network and push out a message with its id and a connect format
// - On receipt, the external nodes should start a listen and write set of channels on the received network address
// - External nodes should then send a message back with the group ip address they're on
// - The new node should then start its own write and listen runners on the broadcast channel. 
// - Using the messages received from each existing node, determine the recurrance counts for each address
// - Config should include a pool limit
// - Node should find a pool that has space and join the group
// - On group join, write a connect message to the group

// type NodeMessageType int64

// const (
// 	Ping NodeMessageType = 0
// 	Connect NodeMessageType = 1
// 	Disconnect NodeMessageType = 2
// 	Message NodeMessageType = 3
// )

// const (
// 	broadcastNet = "224.0.0.1:9999"
// )

// type Msg map[string]string

// func NewMsg(id string, fields map[string]string) []byte {
// 	mx := make(Msg)
// 	mx["id"] = id

// 	for mk, mxx := range fields {
// 		mx[mk] = mxx
// 	}

// 	encd, _ := json.Marshal(mx)
	
// 	return encd
// }

// type Node struct {
// 	Id string
// 	Intf net.Interface
// 	Conn *net.PacketConn
// 	IPV4 *ipv4.PacketConn
// 	Port string
// 	Broadcast *BroadcastLink
// 	ConnectChan chan ConnectMessage
// 	GIP net.IP
// }


// type BroadcastLink struct {
// 	Addr *net.UDPAddr
// 	Conn *net.UDPConn
// 	Write chan []byte
// }

// type ConnectMessage struct {
// 	IP net.IP
// 	Port int
// }

// func NewBroadcastLink(a string) *BroadcastLink {
// 	addr, err := net.ResolveUDPAddr("udp", a)
// 	if err != nil {
// 		log.Printf("new broadcast link err: %v\n", err)
// 		return nil
// 	}

// 	conn, err := net.DialUDP("udp", nil, addr)
// 	if err != nil {
// 		log.Printf("connect new broadcast link err: %v\n", err)
// 		return nil
// 	}

// 	return &BroadcastLink{
// 		Conn: conn,
// 		Write: make(chan []byte, 10),
// 		Addr: addr,
// 	}
// }

// func (N *Node) RunBroadcast() {
// 	go func() {
// 		log.Printf("BroadcastAddr: %v\n", N.Broadcast.Addr)
// 		l, err := net.ListenMulticastUDP("udp", nil, N.Broadcast.Addr)
// 		if err != nil {
// 			log.Printf("broadcast link listen err: %v\n", err)
// 			return
// 		}
// 		l.SetReadBuffer(2500)
// 		for {
// 			b := make([]byte, 2500)
// 			n, src, err := l.ReadFromUDP(b)
// 			if err != nil {
// 				log.Printf("broadcast read err: %v\n", err)
// 				continue
// 			}

			

// 			var msg Msg 
// 			json.Unmarshal(b[:n], &msg)
// 			if val, ok := msg["id"]; ok {
// 				if val == N.Id {
// 					continue
// 				}
// 			}
// 			log.Printf("%v bytes from %v: %v\n", n, src, string(b[:n]))
// 			if val, ok := msg["type"]; ok {
// 				if val == "connect" {
// 					log.Printf("src: %v\n", src)
// 					if nip, ok := msg["group"]; ok {
// 						log.Printf("nip: %v\n", nip)
// 						parsedIP := net.ParseIP(nip)
// 						if np, ok := msg["port"]; ok {
// 							prt, _ := strconv.Atoi(np)

// 							N.ConnectChan <- ConnectMessage{
// 								IP: parsedIP,
// 								Port: prt,
// 							}
// 						}
// 					}
// 					// directConn, err := net.DialUDP("udp", nil, src)
// 					// if err != nil {
// 					// 	log.Print(err)
// 					// 	continue
// 					// }
// 					// _, err := N.IPV4.WriteTo([]byte(NewMsg(N.Id, map[string]string{
// 					// 	"port": N.Port,
// 					// 	"group": N.GIP.String(),
// 					// 	"type": "response",
// 					// })), nil, src)
// 					// if err != nil {
// 					// 	log.Printf("connect resp err: %v\n", err)
// 					// }
// 					// directConn.Write()
					
// 					// b.Write <- []byte(NewMsg(N.Id, map[string]string{
// 					// 	"port": N.Port,
// 					// 	"group": N.GIP.String(),
// 					// }))
// 				}
// 			}
// 			// if nip, ok := msg["group"]; ok {
// 			// 	log.Printf("nip: %v\n", nip)
// 			// 	parsedIP := net.ParseIP(nip)
// 			// 	if np, ok := msg["port"]; ok {
// 			// 		prt, _ := strconv.Atoi(np)

// 			// 		nodeChan <- ConnectMessage{
// 			// 			IP: parsedIP,
// 			// 			Port: prt,
// 			// 		}
// 			// 	}
// 			// }
// 		}

// 	}()

// 	for msg := range N.Broadcast.Write {
// 		N.Broadcast.Conn.Write(msg)
// 	}
// }
// // type NodeMessage struct {
// // 	MType NodeMessageType
// // }

// // type NodeConnectMessage struct {
// // 	GroupIP string

// // }


// func main() {
// 	N := NewNode()
// 	go N.RunBroadcast()
// 	go N.HandleConnectChan()
// 	go N.PacketListen()
// 	N.GIP = net.IPv4(224, 0, 0, 250)

// 	log.Printf("Packet: %v\n", N.IPV4.LocalAddr())
// 	N.Broadcast.Write <- []byte(NewMsg(N.Id, map[string]string{
// 		"type": "connect",
// 	}))
// 	// group := 
// 	N.Broadcast.Write <- []byte(NewMsg(N.Id, map[string]string{
// 		"port": N.Port,
// 		"group": N.GIP.String(),
// 		"type": "connect",
// 	}))
// 	// group := net.IPv4(224, 0, 0, 250)
// 	// if err := N.IPV4.JoinGroup(&N.Intf, &net.UDPAddr{IP: group}); err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// if err := N.IPV4.SetControlMessage(ipv4.FlagDst, true); err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// log.Printf("Group: %v\n", group)
// 	writeChannel := make(chan net.Addr, 10)
// 	// // var groupPort
// 	// go func() {
// 	// 	b := make([]byte, 2500)
// 	// 	for {
// 	// 		_, cm, src, err := N.IPV4.ReadFrom(b)

// 	// 		if err != nil {
// 	// 			log.Printf("Err: %v\n", err)
// 	// 			continue
// 	// 		}

// 	// 		if cm.Dst.IsMulticast() {
// 	// 			if cm.Dst.Equal(group) {
// 	// 				log.Printf("Correct Group: %v\n", string(b))
// 	// 			} else {
// 	// 				log.Printf("From Source: %v\n", src)
// 	// 			}
// 	// 		} else {
// 	// 			log.Printf("Non Multicast: %v\n", string(b))
// 	// 		}
// 	// 		writeChannel <- src
// 	// 	}
// 	// }()

// 	// // convd, _ := strconv.Atoi(N.Port)

// 	// dst := &net.UDPAddr{IP: group}

// 	// if err := N.IPV4.SetMulticastInterface(&N.Intf); err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	// N.IPV4.SetMulticastTTL(2)

// 	// if _, err := N.IPV4.WriteTo([]byte("hello"), nil, dst); err != nil {
// 	// 	log.Printf("send err\n")
// 	// 	log.Fatal(err)
// 	// }

// 	for msg := range writeChannel {
// 		log.Printf("got message: %v\n", msg)
// 		N.IPV4.WriteTo([]byte("ack"), nil, msg)
// 		// N.IPV4.WriteTo([]byte("ack"), nil, msg)
// 	}
	
// }

// func NewNode() *Node {
// 	node := &Node{
// 		Id: uuid.NewString(),
// 		ConnectChan: make(chan ConnectMessage, 10),
// 	}
// 	var castIntf net.Interface
// 	itfs, err := net.Interfaces()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for it := range itfs {
// 		addrs, _ := itfs[it].Addrs()
// 		if itfs[it].Flags&net.FlagMulticast != 0 && 
// 		itfs[it].HardwareAddr != nil && addrs != nil {
// 			castIntf = itfs[it]
// 		}
// 	}
// 	node.Intf = castIntf

// 	c, err := net.ListenPacket("udp4", "0.0.0.0:0")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	addrSplit := strings.Split(c.LocalAddr().String(), ":")
// 	log.Printf("ADDR: %v\n", addrSplit)
// 	node.Port = addrSplit[1]
// 	node.IPV4 = ipv4.NewPacketConn(c)
// 	node.Conn = &c
// 	node.Broadcast = NewBroadcastLink(broadcastNet)
// 	return node
// }
// func (N *Node) HandleConnectChan() {
// 	for cmsg := range N.ConnectChan {
// 		log.Printf("cmsg: %v\n", cmsg)
// 		if err := N.IPV4.JoinGroup(&N.Intf, &net.UDPAddr{IP: cmsg.IP}); err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Printf("group connected\n")
// 		N.IPV4.WriteTo([]byte("ack"), nil, &net.UDPAddr{IP: cmsg.IP})
// 		// N.writeChannel <- &net.UDPAddr{IP: cmsg.IP}
// 	}
// }
// func (N *Node) PacketListen() {
// 	b := make([]byte, 1500)

// 	for {
// 		n, _, src, err := N.IPV4.ReadFrom(b)
// 		if err != nil {
// 			log.Printf("packet listen err: %w\n", err)
// 			continue
// 		}
// 		var msg Msg 
// 		json.Unmarshal(b[:n], &msg)
// 		if val, ok := msg["id"]; ok {
// 			if val == N.Id {
// 				continue
// 			}
// 		}
// 		// if val, ok := msg["type"] == ""
// 		log.Printf("%v bytes received from %v: %v\n", n, src, string(b[:n]))
// 	}
// }