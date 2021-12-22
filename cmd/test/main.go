package main

import (
	"net"
)

func main() {
	nx := NewNode("224.0.0.1:9999")
	// nx.RunBroadcastNet()
	brd := NewGroup("broadcast", net.IPv4(224, 0, 0, 1), 9999)
	nx.Broadcast = &brd
	ConnectToGroup(*nx.Broadcast, nx.Intf)
	go nx.Broadcast.Stream(nx.ReadStream, nx.Id)
	testID := nx.NewGroup("", net.IPv4(225, 0, 0, 1), 0)
	nx.ConnectToGroup(testID)

	// nx.NewGroup("broadcast", net.IPv4(224, 0, 0, 1), 9999)
	// go nx.Groups["broadcast"].Stream(nx.ReadStream)

	// go func() {
	// 	testTicker := time.NewTicker(1 * time.Second)
	// 	for range testTicker.C {
	// 		nx.Broadcast.WriteStream <- []byte(
	// 			NodeMsg{
	// 				Msg: map[string]string{
	// 					"id":      nx.Id,
	// 					"groupid": "broadcast",
	// 					// "string": "i'm a test message",
	// 				},
	// 			}.ToJson(),
	// 		)
	// 	}
	// }()

	go func() {
		for newConn := range nx.ConnectStream {
			go nx.Groups[newConn].Stream(nx.ReadStream, nx.Id)
			nx.Groups[newConn].WriteStream <- []byte(
				NodeMsg{
					Msg: map[string]string{
						"id":      nx.Id,
						"groupid": newConn,
						"msg":     "joined group",
					},
					Dst: nx.Groups[newConn].Addr,
				}.ToJson(),
			)
		}
	}()
	go nx.GroupInitWatch()
	nx.ReadStreamHandler()

}
