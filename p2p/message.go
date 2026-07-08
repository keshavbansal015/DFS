package p2p

const (
	// IncomingMessage indicates that the incoming payload is a standard RPC message.
	IncomingMessage = 0x1
	// IncomingStream indicates that the incoming payload is a raw data stream.
	IncomingStream = 0x2
)

// RPC represents any arbitrary data that is sent over each transport
// between two nodes in the network.
type RPC struct {
	// From holds the network address of the sender.
	From string
	// Payload holds the raw bytes of the message content.
	Payload []byte
	// Stream indicates whether this message is a stream rather than a regular message.
	Stream bool
}
