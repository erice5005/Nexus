package main

// import (
// 	"log"
// 	"net"
// 	"fmt"
// 	"bufio"
// 	"os"
// 	"encoding/json"
// 	// "encoding/base64"
// 	"github.com/google/uuid"
// 	"golang.org/x/net/ipv4"

// )

// //Create unique id
// //Resolve addr
// //Create listenpacket
// //Create basline group (hardcoded)
// //Join group
// //Start listener
// //Broadcast uid and greeting to group

// type Node struct {
// 	Uid string
// 	Addr *net.UDPAddr
// 	Conn net.PacketConn
// 	Packet *ipv4.PacketConn
// 	writeChannel chan NodeMessageFrame
// 	BroadcastChannel chan NodeMessageFrame
// 	Groups map[string]CastGroup
// }

// type NodeMessageFrame struct {
// 	Dest net.Addr
// 	Message NodeMessage
	
// }

// type NodeMessage struct {
// 	MessageType NodeMessageType
// 	Message string
// 	SourceID string
// }

// type CastGroup struct {
// 	Id string
// 	IP *net.UDPAddr
// }

// const (
// 	defaultMulticastAddress = "239.0.0.0:9999"
// )

// func nodeversion() {
// 	newNode, err := NewNode("")
// 	if err != nil {
// 		panic(err)
// 	}

// 	newNode.AssignGroup(CastGroup{
// 		IP: &net.UDPAddr{IP: net.IPv4(224, 0, 0, 250)},
// 		Id: "root",
// 	})

// 	go newNode.Write()
// 	go newNode.WriteBroadcast()
// 	go newNode.ListenMulticast()
// 	newNode.BroadcastChannel <- NodeMessageFrame{
// 		Message: NodeMessage{
// 			SourceID: newNode.Uid,
// 			MessageType: Connect,
// 			Message: fmt.Sprintf("hello from: %s", newNode.Uid),
// 		},
// 	}
// 	go newNode.Listen()
//     scanner := bufio.NewScanner(os.Stdin)
    
//     for {
//         fmt.Print("Enter message: ")
        
//         scanner.Scan()
        
//         text := scanner.Text()

//         if len(text) != 0 {

// 			// newAddr, err := net.ResolveUDPAddr("udp4", text)
// 			// if err != nil {
// 			// 	log.Printf("err: %v\n", err)
// 			// }
// 			newNode.writeChannel <- NodeMessageFrame{
// 				// Dest: newNode.Groups["root"].IP,
// 				Message: NodeMessage{
// 					SourceID: newNode.Uid,
// 					MessageType: Connect,
// 					Message: text,
// 				},
// 			}
//         } else {
//             break
//         }
//     }
// }

// func NewNode(address string) (*Node, error) {
// 	if address == "" {
// 		address = defaultMulticastAddress
// 	}
// 	addr, err := net.ResolveUDPAddr("udp4", address)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Printf("Addr: %v\n", addr.String())
// 	c, err := net.ListenPacket("udp4", addr.String())
// 	if err != nil {
// 		return nil, err
// 	}

// 	p := ipv4.NewPacketConn(c)
// 	en0, err := net.InterfaceByName("eth0")
// 	if err != nil {
// 		log.Fatal(err)
// 	}	
// 	addrs, _ :=  en0.Addrs()
// 	log.Printf("EN: %v\n",addrs)
// 	p.SetMulticastInterface(en0)
// 	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
// 		// error handling
// 		log.Println(err)
// 	}
// 	p.SetMulticastTTL(2)
// 	return &Node{
// 		Uid: uuid.NewString(),
// 		Addr: addr, 
// 		Conn: c, 
// 		Packet: p,
// 		writeChannel: make(chan NodeMessageFrame, 10),
// 		Groups: make(map[string]CastGroup),
// 		BroadcastChannel: make(chan NodeMessageFrame, 10),
// 	}, nil
// }

// func (n *Node) Listen() error {
// 	// var err error
// 	b := make([]byte, 2500)
// 	for {
// 		nc, _, src, err := n.Packet.ReadFrom(b)
// 		if err != nil {
// 			return err
// 		}
// 		// log.Printf("Bytes: %v from %v: %v\n", nc, src, string(b))

// 		recvd := NodeMessage{}
// 		err = json.Unmarshal(b[:nc], &recvd)
// 		if err != nil {
// 			log.Println(err)
// 			// return err
// 			continue
// 		}
// 		if recvd.SourceID == n.Uid {
// 			continue
// 		}
// 		// recvd.Message, _ = base64.StdEncoding.DecodeString(string(recvd.Message))
// 		log.Printf("Recieved %v from %v\n", recvd, src)
// 		n.writeChannel <- NodeMessageFrame{
// 			Dest: src,
// 			Message: NodeMessage{
// 				Message: recvd.Message,
// 			},
// 		}
// 	}
// }

// func (n *Node) ListenMulticast() error {
// 	conn, err := net.ListenMulticastUDP("udp4", nil, n.Addr)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("ConnAddr: %v\n", conn.LocalAddr().(*net.UDPAddr))
// 	b := make([]byte, 2500)
// 	for {
// 		nc, src, err := conn.ReadFromUDP(b)
// 		// nc, _, src, err := n.Packet.ReadFrom(b)
// 		if err != nil {
// 			log.Fatal(err)
// 			return err
// 		}
	
		

	

// 		recvd := NodeMessage{}
// 		err = json.Unmarshal(b[:nc], &recvd)
// 		if err != nil {
// 			log.Println(err)
// 			// return err
// 			continue
// 		}

// 		if recvd.SourceID == n.Uid {
// 			continue
// 		}


// 		log.Printf("Broadcast: %v from %v: %v\n", nc, src, string(b))

// 		n.writeChannel <- NodeMessageFrame{
// 			Dest: src,
// 			Message: recvd,
// 		}
// 	}
// }

// func (n *Node) Write() error {
// 	for nm := range n.writeChannel {
// 		if nm.Dest == nil {
// 			nm.Dest = n.Groups["root"].IP
// 		}
// 		nm.Message.SourceID = n.Uid
// 		encd, _  := json.Marshal(nm.Message)
// 		n.Packet.WriteTo(encd, nil, nm.Dest)
// 	}

// 	return nil
// }

// func (n *Node) WriteBroadcast() error {
// 	conn, err := net.DialUDP("udp4", nil, n.Addr)
// 	if err != nil {
// 		return err
// 	}
// 	for nm := range n.BroadcastChannel {
// 		nm.Message.SourceID = n.Uid
// 		encd, _  := json.Marshal(nm.Message)
// 		conn.Write(encd)
// 	}

// 	return nil
// }

// func (n *Node) AssignGroup(cg CastGroup) {
// 	en0, err := net.InterfaceByName("eth0")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if _, ok := n.Groups[cg.Id]; !ok {
// 		log.Printf("Joining: %v\n", cg.Id)
// 		n.Groups[cg.Id] = cg
// 		err = n.Packet.JoinGroup(en0, cg.IP)
// 		if err != nil {
// 			log.Printf("join err :%v\n", err)
// 			return
// 		}
// 	}
// }