package nexus

import "net"

func NewNexusNode() *Node {
	nx := NewNode("224.0.0.1:9999")
	brd := NewGroup("broadcast", net.IPv4(224, 0, 0, 1), 9999)
	nx.Broadcast = &brd

	return nx
}

func RunNexusNode(nx *Node) {
	ConnectToGroup(*nx.Broadcast, nx.Intf)
	go nx.Broadcast.Stream(nx.ReadStream, nx.Id)
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
