package p2p

import "net"

// HandshakeFunc is a function type representing a connection handshake protocol
// that is executed upon establishing a connection before network operations begin.
type HandshakeFunc func(net.Conn) error

// NOPHandshakeFunc is a default no-op implementation of HandshakeFunc that returns nil.
func NOPHandshakeFunc(net.Conn) error { return nil }

// TODO: Implement an actual handshake function.
