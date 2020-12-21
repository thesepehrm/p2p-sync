package main

import (
	"flag"
	"fmt"
	"log"

	"gitlab.com/thesepehrm/p2p-sync/p2p"
)

func main() {

	help := flag.Bool("h", false, "Display Help")
	port := flag.Int("p", 2000, "Port number of the node")
	dest := flag.String("d", "", "destination address")

	flag.Parse()

	if *help {
		fmt.Println("A simple chat p2p application")
		fmt.Println()
		flag.PrintDefaults()
		return
	}

	node := p2p.NewNode(*port)

	if *dest != "" {
		err := node.Connect(*dest)
		if err != nil {
			log.Panic(err)
		}
	} else {

		node.Start()

	}

	// Run forever
	<-make(chan struct{})
}
