package main

import (
	"github.com/google/uuid"
	"net"
	"log"
	"time"
	"encoding/json"
	"strconv"
	"strings"
	"golang.org/x/net/ipv4"
)

const (
	broadcastNet = "224.0.0.1:9999"
	maxDatagramSize = 8192
)

var nodeid string

type NodeMsg struct {
	Dst *net.UDPAddr
	Msg map[string]string
}

var seenNodes map[string]time.Time

var groupIP net.IP
var groupPort int
var groupWriteChan chan []byte
var directWrite chan NodeMsg
var pingChan chan NodeMsg
var firstMsg time.Time
func (nm NodeMsg) ToJson() string {
	encd, err := json.Marshal(nm)
	if err != nil {
		return ""
	}
	return string(encd)
}

type targetedMsg struct {
	Dst *net.UDPAddr
	Msg NodeMsg
}

func MsgFromJson(inp []byte) NodeMsg {
	var out NodeMsg
	err := json.Unmarshal(inp, &out)
	if err != nil {
		return NodeMsg{}
	}
	return out
}

func main() {
	nodeid = uuid.NewString()
	groupWriteChan = make(chan []byte, 10)
	directWrite = make(chan NodeMsg, 10)
	seenNodes = make(map[string]time.Time)
	pingChan = make(chan NodeMsg, 10)
	targetAddr, _ := net.ResolveUDPAddr("udp", broadcastNet)
	groupPort = 0
	go ping(targetAddr)
	go readMulticast(targetAddr, handler)
	// ticker := time.NewTicker(2 * time.Second)
	
	time.Sleep(2 * time.Second) 
	eth0, err := net.InterfaceByName("eth0")
	if err != nil {
		log.Printf("interface err: %v\n", err)
	}
	log.Printf("Seen count: %v\n", len(seenNodes))
	if firstMsg.IsZero() {
		time.Sleep(2 * time.Second)
		log.Printf("%v\n",firstMsg)
	}
	if len(seenNodes) == 0 {
		groupIP = net.IPv4(224, 0, 0, 250)	
	}
	// time.Sleep(2 * time.Second)
	c, err := net.ListenPacket("udp", net.JoinHostPort("224.0.0.1", "0"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("C: %v\n", c)
	p := ipv4.NewPacketConn(c)
	if groupPort == 0 {
		hostPortVal := p.LocalAddr()

	hostSplit := strings.Split(hostPortVal.String(), ":")
	if len(hostSplit) == 2 {
		convd, _ := strconv.Atoi(hostSplit[1])
		groupPort = convd
	}
	}
	
	log.Printf("Group IP: %v, Port: %v\n", groupIP, groupPort)
	if err := p.JoinGroup(eth0, &net.UDPAddr{IP: groupIP, Port: groupPort}); err != nil {
		// error handling
		log.Printf("group connect err: %v\n", err)
	} else {
		groupWriteChan <- []byte(NodeMsg{Msg: map[string]string{"id": nodeid}}.ToJson())
	}
	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
		// error handling
		log.Printf("group connect err: %v\n", err)

	}
	go func() {
		b := make([]byte, 1500)
		for {
			_, cm, src, err := p.ReadFrom(b)
			if err != nil {
				continue
				// error handling
			}
			if src == p.LocalAddr() {
				continue
			}
			if cm.Dst == nil {
				continue
			}
			if cm.Dst.IsMulticast() {
				if cm.Dst.Equal(groupIP) {
					log.Printf("Group: %v\n", cm)
					// joined group, do something
				} else {
					// unknown group, discard
					continue
				}
			}
		}
	}()
	go func() {
		if err := p.SetMulticastInterface(eth0); err != nil {
			// error handling
			log.Printf("write err: %v\n", err)

		}
		for msg := range groupWriteChan {
			if _, err := p.WriteTo(msg, nil, &net.UDPAddr{IP: groupIP, Port: groupPort}); err != nil {
				log.Printf("write err: %v\n", err)
			}
		}
		log.Printf("end write loop\n")
	}()
	for msg := range directWrite {
		log.Printf("Direct: %v\n", msg)
		if _, err := p.WriteTo([]byte(msg.ToJson()), nil, msg.Dst); err != nil {
			log.Printf("write err: %v\n", err)
		}
	}
}

func readMulticast(addr *net.UDPAddr, h func(*net.UDPAddr, int, []byte)) {
	l, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Printf("listen err: %v\n", err)
	}
	l.SetReadBuffer(maxDatagramSize)

	for {
		b := make([]byte, maxDatagramSize)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Printf("read err: %v\n", err)
		}
		
		h(src, n, b)
	}
}

func ping(a *net.UDPAddr) {
	c, err := net.DialUDP("udp",nil, a)
	if err != nil {
		log.Printf("dial err:%v\n", err)
	}

	go func() {
		for msg := range pingChan {
			c.Write([]byte(msg.ToJson()))
		}
	}()
	for {
		c.Write([]byte(NodeMsg{Msg: map[string]string{"id": nodeid}}.ToJson()))
		time.Sleep(1 * time.Second)
	}
}

func handler(src *net.UDPAddr, n int, b []byte) {
			// log.Printf("%v bytes from %v: %v", n, src, string(b[:n]))

	msg := MsgFromJson(b[:n])
	if _, ok := msg.Msg["id"]; !ok {
		return
	}

	srcId := msg.Msg["id"]
	// log.Printf("DST: %v\n", msg.Dst)
	if srcId == nodeid {
		// log.Printf("%v bytes from %v: %v", n, src, string(b[:n]))

		return
	}
	firstMsg = time.Now()
	// canSetIP := false
	// sendConnect := false
	// log.Printf("MSG: %v\n", msg)
	if _, ok := seenNodes[srcId]; !ok {
		// groupWriteChan <- 
		// if groupIP == nil {
			// canSetIP = true
			// sendConnect = true
		if groupIP != nil {
			log.Printf("can send: %v\n", groupIP)
			pingChan <- NodeMsg{Msg: map[string]string{"id": nodeid, "groupIP": groupIP.String(), "groupPort": strconv.Itoa(groupPort)}}
		}
		// seenNodes[srcId] = time.Now()
	// } else {
		
		
	}
	if groupIP == nil {
		log.Printf("%v\n", msg)
		if val, ok := msg.Msg["groupPort"]; ok {
			conv, _ := strconv.Atoi(val)
			log.Printf("Conv: %v\n", conv)
			groupPort = conv
		}
		if val, ok := msg.Msg["groupIP"]; ok {
			// if canSetIP {
				groupIP = net.ParseIP(val)
				// groupIP = src.IP
			// }
		}
	}
	seenNodes[srcId] = time.Now()

	
	log.Printf("Msg from :%v", srcId)
}