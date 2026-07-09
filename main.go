// package main implements the entrypoint and server bootstrap helpers for the distributed file system (DFS).
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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

// main launches three DFS server instances on ports 3000, 7005, and 5005 respectively,
// bootstraps their network topology, and starts an interactive CLI shell
// allowing you to store, delete, and retrieve files over the network.
func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":7005", "")
	s3 := makeServer(":5005", ":3000", ":7005")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s3.Start()) }()

	// Wait for topology connections to settle
	time.Sleep(2 * time.Second)

	fmt.Println("\n=======================================================")
	fmt.Println("       Distributed File System (DFS) Interactive Demo")
	fmt.Println("=======================================================")
	fmt.Println("Nodes Booted:")
	fmt.Printf(" - Node 1 on port :3000 (bootstrap)\n")
	fmt.Printf(" - Node 2 on port :7005 (bootstrap)\n")
	fmt.Printf(" - Node 3 on port :5005 (connected to Node 1 and Node 2)\n")
	fmt.Println("You will interact with Node 3 (:5005) through this CLI.")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  store <key> <content>  - Encrypts & stores file on Node 3 and replicates it to Node 1 & 2")
	fmt.Println("  delete <key>           - Deletes the local copy of the file on Node 3 (forces a network lookup on next get)")
	fmt.Println("  get <key>              - Retrieves and displays the file content (will fetch from Node 1/2 if deleted locally)")
	fmt.Println("  help                   - Shows the list of commands")
	fmt.Println("  exit / quit            - Exits the application")
	fmt.Println("=======================================================\n")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("DFS> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 3)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "exit", "quit":
			fmt.Println("Exiting. Goodbye!")
			return

		case "help":
			fmt.Println("Commands: ")
			fmt.Println("  store <key> <content>")
			fmt.Println("  delete <key>")
			fmt.Println("  get <key>")
			fmt.Println("  exit")

		case "store":
			if len(parts) < 3 {
				fmt.Println("Usage: store <key> <content>")
				continue
			}
			key := parts[1]
			content := parts[2]

			data := bytes.NewReader([]byte(content))
			err := s3.Store(key, data)
			if err != nil {
				fmt.Printf("Error storing file: %s\n", err)
			} else {
				fmt.Printf("Successfully stored key '%s' on s3 and replicated across network\n", key)
			}

		case "delete":
			if len(parts) < 2 {
				fmt.Println("Usage: delete <key>")
				continue
			}
			key := parts[1]
			err := s3.store.Delete(s3.ID, key)
			if err != nil {
				fmt.Printf("Error deleting local file: %s\n", err)
			} else {
				fmt.Printf("Successfully deleted local copy of '%s' from s3's disk\n", key)
			}

		case "get":
			if len(parts) < 2 {
				fmt.Println("Usage: get <key>")
				continue
			}
			key := parts[1]

			r, err := s3.Get(key)
			if err != nil {
				fmt.Printf("Error retrieving file: %s\n", err)
				continue
			}

			b, err := io.ReadAll(r)
			if err != nil {
				fmt.Printf("Error reading file stream: %s\n", err)
				continue
			}

			fmt.Printf("File Content for '%s':\n>>> %s <<<\n", key, string(b))

		default:
			fmt.Printf("Unknown command: %s. Type 'help' for options.\n", cmd)
		}
	}
}
