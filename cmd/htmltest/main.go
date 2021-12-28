package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	nexus "github.com/erice8996/Nexus"
)

var activeNodes []*nexus.Node

func main() {
	activeNodes = make([]*nexus.Node, 0)
	// activeNodes = append(activeNodes, n1)
	// time.Sleep(1 * time.Second)
	commonId := "commongroup"
	// go func() {
	for i := 0; i < 250; i++ {
		nn := nexus.NewNode(nexus.NodeConfig{}, fmt.Sprintf("node %v", i))

		nn.ConnectGroup(nexus.NewGroup(
			commonId,
			nexus.NexusGroupAddress{
				IP:   net.IPv4(224, 0, 0, 1),
				Port: 1234,
			},
		))

		if i%2 == 0 {
			evenGroup := nexus.NewGroup("even", nexus.NexusGroupAddress{
				IP:   net.IPv4(224, 0, 0, 4),
				Port: 2222,
			})
			evenGroup.Private = true
			nn.ConnectGroup(evenGroup)
		}
		if i%5 == 0 {
			evenGroup := nexus.NewGroup("five", nexus.NexusGroupAddress{
				IP:   net.IPv4(224, 0, 0, 4),
				Port: 5555,
			})
			evenGroup.Private = true
			nn.ConnectGroup(evenGroup)
		}
		nn.Run()
		log.Printf("Running: %v\n", nn.Id)
		// go func() {
		// for _, g := range nn.Groups {
		// 	go func() {
		// 		for msg := range g.Input {
		// 			log.Printf("Node: %v, Group: %v, Message: %v\n", nn.Id, g.Group.Id, msg)
		// 		}
		// 	}()
		// }
		activeNodes = append(activeNodes, nn)
		// }
		// time.Sleep(500 * time.Millisecond)
	}
	// }()
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for c := range ticker.C {
			log.Printf("C: %v\n", c)
			log.Printf("Ticker\n")
			log.Printf("nodes: %v\n", activeNodes)
			for _, an := range activeNodes {
				log.Printf("AN: %v, group length: %v\n", an.Id, len(an.Groups))
				for _, g := range an.Groups {

					g.Output <- nexus.NodeMessage{
						SenderId: an.Id,
						Routing: nexus.MessageRouting{
							MessageType:        nexus.Message,
							DestinationAddress: g.Group.Address.Address,
						},
						Data: map[string]interface{}{
							"SentFrom": an.Id,
						},
					}
					log.Printf("Group ID: %v\n", g.Group.Id)
				}
				log.Printf("complete: %v\n", an.Id)
			}
			log.Printf("Finished run: %v\n", c)
		}
		log.Panicf("done\n")
	}()

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		out := make([]string, 0)
		for _, nx := range activeNodes {
			// b, err := json.MarshalIndent(nx.Network, "", "  ")
			// if err != nil {
			// 	fmt.Println("error:", err)
			// }
			out = append(out, fmt.Sprintf("Id: %v, Seen: %v, Groups: %v\n", nx.Id, len(nx.Network), len(nx.Groups)))
		}
		fmt.Fprint(rw, strings.Join(out, "\n"))
	})

	http.ListenAndServe("localhost:8080", nil)
	// time.Sleep(1 * time.Minute)
}
