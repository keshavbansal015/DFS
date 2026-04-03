package main

import (
	"fmt"
	"log"

	"github.com/keshavbansal015/DFS/p2p"
)

func main() {

	tr := p2p.NewTCPTransport(":3000")

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("We Gucci")
	select {}
}
