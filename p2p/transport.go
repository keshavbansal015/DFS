package p2p

import "net"

// Peer represents a remote node in the network established over a connection.
type Peer interface {
	net.Conn
	// Send writes a slice of bytes to the peer.
	Send([]byte) error
	// CloseStream signals that a stream is finished copying/reading.
	CloseStream()
}

// Transport manages the physical connection protocol between nodes in the network
// (e.g., TCP, UDP, WebSockets).
type Transport interface {
	// Addr returns the local listener network address.
	Addr() string
	// Dial connects to a remote network address.
	Dial(string) error
	// ListenAndAccept binds to the configured port and starts accepting incoming connections.
	ListenAndAccept() error
	// Consume returns a read-only channel containing parsed RPC messages received from peers.
	Consume() <-chan RPC
	// Close terminates the transport listener and releases resources.
	Close() error
}
