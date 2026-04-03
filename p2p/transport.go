package p2p

// Represents remote node
type Peer interface {
}

// Anything that handles the communication between,
// the nodes in the network
// (TCP, UDP, WebSockets, ...)
type Transport interface {
	ListenAndAccept() error
}
