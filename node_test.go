package nexus

import (
	"log"
	"net"
	"testing"
	"time"
)

func Test_NodeBroadcast(t *testing.T) {
	testNode := NewNode(NodeConfig{
		Groups: -1,
	}, "")

	broadcastNet := NewGroup("broadcast", NexusGroupAddress{
		IP:   net.IPv4(224, 0, 0, 250),
		Port: 9999,
	})
	if broadcastNet.Connection.IPV4 == nil {
		t.Fail()
	}

	testNode.BroadcastGroup = ConnectedNexusGroup{
		Group:        broadcastNet,
		Output:       make(chan NodeMessage, 10),
		Input:        make(chan interface{}, 10),
		InputHandler: testNode.BroadcastInputHandler,
	}

	testNode.BroadcastGroup.Group.Connect(testNode.Config.Intf)

	go func() {
		testNode.HandleBroadcastStream()
	}()
	go testNode.HandleConnectStream()
	time.Sleep(25 * time.Second)
}

func Test_TwoNodes(t *testing.T) {
	testNode1 := NewNode(NodeConfig{
		Groups: -1,
	}, "")

	testNode2 := NewNode(NodeConfig{
		Groups: -1,
	}, "")

	if testNode2.BroadcastGroup.Group.Connection.IPV4 == nil {
		t.Fail()
	}

	testNode1.Run()
	time.Sleep(5 * time.Second)

	testNode2.Run()
	time.Sleep(25 * time.Second)

}

func Test_NodeExclusion(t *testing.T) {
	testNode1 := NewNode(NodeConfig{
		Groups: -1,
	}, "node1")

	testNode2 := NewNode(NodeConfig{
		Groups: -1,
	}, "node2")

	testNode3 := NewNode(NodeConfig{
		Groups: -1,
	}, "node3")

	oneAndTwoGroup := NewGroup("1+2", NexusGroupAddress{
		IP:   net.IPv4(224, 0, 0, 1),
		Port: 4200,
	})
	oneAndTwoGroup.Private = true
	oneAndTwoGrouptwo := NewGroup("1+2", NexusGroupAddress{
		IP:   net.IPv4(224, 0, 0, 1),
		Port: 4200,
	})
	oneAndTwoGrouptwo.Private = true
	threeSoloGroup := NewGroup("3", NexusGroupAddress{
		IP:   net.IPv4(224, 0, 0, 2),
		Port: 6900,
	})
	threeSoloGroup.Private = true
	threeSoloGroupTwo := NewGroup("3", NexusGroupAddress{
		IP:   net.IPv4(224, 0, 0, 2),
		Port: 6900,
	})
	threeSoloGroupTwo.Private = true
	testNode1.Run()
	log.Println("Node 1 started")
	// time.Sleep(1 * time.Second)
	testNode2.Run()
	// time.Sleep(1 * time.Second)
	testNode3.Run()
	testNode1.ConnectGroup(oneAndTwoGroup)
	log.Printf("Connected: %v\n", testNode1)
	testNode2.ConnectGroup(oneAndTwoGrouptwo)
	log.Printf("Connected: %v\n", testNode2)

	testNode3.ConnectGroup(threeSoloGroup)
	testNode2.ConnectGroup(threeSoloGroupTwo)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		for range ticker.C {
			testNode2.Groups[oneAndTwoGroup.Id].Output <- NodeMessage{
				SenderId: testNode2.Id,
				Routing: MessageRouting{
					MessageType:        Message,
					DestinationAddress: oneAndTwoGroup.Address.Address,
				},
				Data: map[string]interface{}{
					"1+2": "I can only be seen by 1+2",
				},
			}
			testNode3.Groups[threeSoloGroup.Id].Output <- NodeMessage{
				SenderId: testNode3.Id,
				Routing: MessageRouting{
					MessageType:        Message,
					DestinationAddress: threeSoloGroup.Address.Address,
				},
				Data: map[string]interface{}{
					"3": "I can only be seen by 2",
				},
			}
		}
	}()

	go func() {
		for msg := range testNode1.Groups[oneAndTwoGroup.Id].Input {
			log.Printf("Test Node 1 received: %v\n", msg)
		}
	}()
	go func() {
		for msg := range testNode2.Groups[oneAndTwoGroup.Id].Input {
			log.Printf("Test Node 2 received: %v\n", msg)
		}
	}()
	go func() {
		for msg := range testNode2.Groups[threeSoloGroup.Id].Input {
			log.Printf("Test Node 2 received: %v\n", msg)
		}
	}()
	go func() {
		for msg := range testNode3.Groups[threeSoloGroup.Id].Input {
			log.Printf("Test Node 3 received: %v\n", msg)
		}
	}()
	log.Printf("Start sleep: %v\n", time.Now())
	time.Sleep(25 * time.Second)
	// testNode1.Groups[oneAndTwoGroup.Id]
	// testNode2.Groups[oneAndTwoGroup.Id] = oneAndTwoGroup
	// testNode3.Groups[threeSoloGroup.Id] = threeSoloGroup

}
