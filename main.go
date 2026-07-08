// package main implements the entrypoint and server bootstrap helpers for the distributed file system (DFS).
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/keshavbansal015/DFS/crypto"
	"github.com/keshavbansal015/DFS/p2p"
	"github.com/keshavbansal015/DFS/store"
)

// makeServer is a helper function that constructs a new FileServer instance configured
// with a TCP transport on the specified listen address, optional bootstrap nodes,
// and default parameters like no-op handshakes and default decoders.
func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := FileServerOpts{
		EncKey:            crypto.NewEncryptionKey(),
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: store.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)

	tcpTransport.OnPeer = s.OnPeer
	return s
}

// main launches three DFS server instances on ports 3000, 7000, and 5000 respectively,
// bootstraps their network topology, and executes a series of store/get cycles
// to demonstrate local and network file storage and retrieval functionality.
func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":7005", "")
	s3 := makeServer(":5005", ":3000", ":7005")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()

	time.Sleep(2 * time.Second)

	go s3.Start()
	time.Sleep(2 * time.Second)

	for i := 0; i < 2; i++ {
		key := fmt.Sprintf("picture_%d", i)
		data := bytes.NewReader([]byte("FILE DATA"))
		s3.Store(key, data)

		if err := s3.store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	}
}
