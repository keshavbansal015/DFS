package p2p

import "net"

// Represents remote node
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Anything that handles the communication between,
// the nodes in the network
// (TCP, UDP, WebSockets, ...)
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
