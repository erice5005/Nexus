package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	nexus "github.com/erice8996/Nexus"
)

var activeNodes []*nexus.Node

func main() {
	activeNodes = make([]*nexus.Node, 0)

	n1 := nexus.NewNexusNode()
	// n1.GroupLimit = 5
	// nx := nexus.NewNode("224.0.0.1:9999")
	// brd := nexus.NewGroup("broadcast", net.IPv4(224, 0, 0, 1), 9999)
	// nx.Broadcast = &brd
	go nexus.RunNexusNode(n1)
	activeNodes = append(activeNodes, n1)
	time.Sleep(1 * time.Second)

	go func() {
		for i := 0; i < 10; i++ {
			nn := nexus.NewNexusNode()
			// nn.GroupLimit = 4
			go nexus.RunNexusNode(nn)
			activeNodes = append(activeNodes, nn)
			time.Sleep(5 * time.Second)
		}
	}()

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		out := make([]string, 0)
		for _, nx := range activeNodes {
			b, err := json.MarshalIndent(nx.Network, "", "  ")
			if err != nil {
				fmt.Println("error:", err)
			}
			out = append(out, fmt.Sprintf("Id: %v, Seen: %v, Groups: %v\n", nx.Id, string(b), len(nx.Groups)))
		}
		fmt.Fprint(rw, strings.Join(out, "\n"))
	})

	http.ListenAndServe("localhost:8080", nil)
	// time.Sleep(1 * time.Minute)
}
