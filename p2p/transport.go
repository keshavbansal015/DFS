package p2p

// Represents remote node
type Peer interface {
	Close() error
}

// Anything that handles the communication between,
// the nodes in the network
// (TCP, UDP, WebSockets, ...)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
}
