package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents the remote node over an established TCP connection.
type TCPPeer struct {
	// Conn is the underlying net.Conn network socket.
	net.Conn
	// outbound is true if this connection was established via outbound Dialing,
	// or false if it was accepted via an inbound Listen.
	outbound bool
	// wg coordinates concurrent stream tasks.
	wg *sync.WaitGroup
}

// NewTCPPeer instantiates a new TCPPeer wrapping the given connection.
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

// CloseStream signals that an active stream copy task has finished.
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

// Send writes the raw byte slice directly to the peer connection socket.
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

// TCPTransportOpts defines the configuration parameters for the TCPTransport.
type TCPTransportOpts struct {
	// ListenAddr is the network address where the TCP server will listen for connections.
	ListenAddr string
	// HandshakeFunc is the logic invoked to validate a connection on setup.
	HandshakeFunc HandshakeFunc
	// Decoder is the decoder interface used to deserialize stream messages.
	Decoder Decoder
	// OnPeer is a callback fired when a new Peer connects and completes the handshake successfully.
	OnPeer func(Peer) error
}

// TCPTransport implements the Transport interface using TCP connections.
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

// NewTCPTransport instantiates a new TCPTransport with the specified options.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
	}
}

// Addr implements the Transport interface, returning the address where the transport accepts connections.
func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

// Consume implements the Transport interface, returning a read-only channel for RPC messages.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close implements the Transport interface, shutting down the TCP listener socket.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial establishes an outbound TCP connection to the specified address.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)
	log.Printf("[DIL] %s => %s\n", t.Addr(), addr)
	return nil
}

// ListenAndAccept binds the TCP listener to the configured address and begins accepting connections.
func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("[LISTENING] => %s\n", t.ListenAddr)

	return nil
}

// startAcceptLoop continuously accepts inbound TCP connections until the listener is closed.
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		go t.handleConn(conn, false)
	}
}

// handleConn manages the connection lifecycle: performing handshakes, invoking callbacks,
// and starting the message read loop.
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}

		// sending all the rpc to the rpcch, basically the main thread will be consuming from it.
		t.rpcch <- rpc
	}
}
