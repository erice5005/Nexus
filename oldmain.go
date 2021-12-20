package main

// import (
// 	"net"
// 	"log"
// 	// "os"
// 	"github.com/google/uuid"
// 	"time"
// 	"golang.org/x/net/ipv4"
// 	// "strconv"
// 	// "encoding/hex"
// 	// "github.com/urfave/cli"
// )

// const (
// 	maxDatagramSize = 8192
// 	// defaultMulticastAddress = "239.0.0.0:9999"
// )

// // type Node struct {
// // 	Id string
// // 	Addr string
// // 	PingRate time.Duration
// // 	IsLeaderNode bool

// // }

// type NodeMessageType int64

// const (
// 	PingMessage NodeMessageType = 0
// 	Connect NodeMessageType = 1
// 	Disconnect NodeMessageType = 2
// 	Message NodeMessageType = 3
// )

// // type NodeMessage struct {

// // }

// // Topology
// // Base: Broadcast net, all nodes join, one creates if none exists
// // Each node maintains a list of all nodes within the broadcast net
// // Each node broadcasts the port on which 



// func oldmain() {	

// 	// First find out if there is already a leader node
// 	uid := uuid.NewString()
// 	en0, err := net.InterfaceByName("eth0")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	listAddr, err := net.ResolveUDPAddr("udp4", defaultMulticastAddress)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	c, err := net.ListenPacket("udp4", listAddr.String())
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer c.Close()

// 	p := ipv4.NewPacketConn(c)
// 	group := net.IPv4(224, 0, 0, 250)

// 	// argsWithoutProg := os.Args[1:]
// 	// port := argsWithoutProg[0]
// 	// inttt, err := net.Interfaces()
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// log.Printf("Interfaces: %v\n", inttt)
// 	// en0, err := net.InterfaceByName("eth0")
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
	

// 	// c, err := net.ListenPacket("udp4", "224.0.0.0:" + port)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	// defer c.Close()

// 	// p := ipv4.NewPacketConn(c)


// 	// var group net.IP
// 	// // if argsWithoutProg[1] {
// 	// // 	existingGroup := argsWithoutProg[1]
// 	// // 	group = 
// 	// // }

// 	// group := net.IPv4(224, 0, 0, 250)
// 	// c.Write([]byte(listAddr.String()))

// 	br, _ := Broadcaster(defaultMulticastAddress)
// 	go Listener(defaultMulticastAddress, handler)
// 	br.Write([]byte(uid))

// 	log.Printf("UID: %v\n, GroupIP: %v\n", uid, group)
// 	if err := p.JoinGroup(en0, &net.UDPAddr{IP: group}); err != nil {
// 		log.Fatal(err)
// 	}
// 	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("C: %v\n", br.LocalAddr())
// 	// go func() {
// 		b := make([]byte, 1500)
// 		for {
// 			n, cm, src, err := p.ReadFrom(b)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			log.Printf("N: %v, SRC: %v, DST: %v, MSG: %v\n", n, src, string(b))
// 			log.Printf("DST: %v\n", cm.Dst.IsMulticast())
// 		}
		
// 	// }()
// 	// portNum, err := strconv.Atoi(port)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	// dst := &net.UDPAddr{IP: group, Port: portNum}
// 	// if err := p.SetMulticastInterface(en0); err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// p.SetMulticastTTL(2)
// 	// for {
// 	// 	if _, err := p.WriteTo([]byte("hello, world"), nil, dst); err != nil {
// 	// 		log.Fatal(err)
// 	// 	}

// 	// 	time.Sleep(1 * time.Second)
// 	// }
// 	// // argsWithoutProg := os.Args[1:]
// 	// // switch argsWithoutProg[0] {
// 	// // case "B":
// 	// // 	Ping(defaultMulticastAddress)
// 	// // case "L":
// 	// // 	Listener(defaultMulticastAddress, func(src *net.UDPAddr, n int, b []byte) {
// 	// // 		// log.Printf("%v bytes from %v\n", n, src)
// 	// // 		// log.Println(hex.Dump(b[:n]))
// 	// // 		log.Printf("%v\n", string(b))
// 	// // 	})
// 	// // }
// 	// // app := cli.NewApp()
	

// 	// log.Printf("TESTTT\n")
// }
// func handler(src *net.UDPAddr, n int, b []byte) {
// 			// log.Printf("%v bytes from %v\n", n, src)
// 			// log.Println(hex.Dump(b[:n]))
// 			log.Printf("Handler: %v\n", string(b))
// 		}
// func Broadcaster(address string) (*net.UDPConn, error) {
// 	addr, err := net.ResolveUDPAddr("udp4", address)
// 	if err != nil {
// 		return nil, err
// 	}

// 	conn, err := net.DialUDP("udp4", nil, addr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return conn, nil
// }

// func Listener(address string, handler func(*net.UDPAddr, int, []byte)) {
// 	addr, err := net.ResolveUDPAddr("udp4", address)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
	
// 	conn.SetReadBuffer(maxDatagramSize)

// 	for {
// 		buffer := make([]byte, maxDatagramSize)
// 		numBytes, src, err := conn.ReadFromUDP(buffer)

// 		if err != nil {
// 			log.Fatal("Read from UDP failed: ", err)
// 		}
// 		log.Printf("Broadcast from: %v\n", src)
// 		handler(src, numBytes, buffer)
// 	}
// }

// func Ping(addr string) {
// 	conn, err := Broadcaster(addr)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for {
// 		conn.Write([]byte("hello, world\n"))
// 		time.Sleep(1 * time.Second)
// 	}
// }