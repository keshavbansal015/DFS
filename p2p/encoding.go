package p2p

import (
	"encoding/gob"
	"io"
)

// Decoder interface for decoding binary data.
type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{} // implements Decoder using gob serialization.

func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

// implements a decoder that supports peek buffering for streams and
// reading raw unformatted message chunks.
type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	peekBuf := make([]byte, 1)

	if _, err := r.Read(peekBuf); err != nil {
		return nil
	}

	stream := peekBuf[0] == IncomingStream

	if stream {
		msg.Stream = true
		return nil
	}

	buf := make([]byte, 1028)

	n, err := r.Read(buf)
	if err != nil {
		return nil
	}

	msg.Payload = buf[:n]

	return nil
}

type Encoder interface {
	Encode(io.Writer, *RPC) error
}

type DefaultEncoder struct{}

// PENDING: implementation.
func (enc DefaultEncoder) Encode(w io.Writer, msg *RPC) error {
	return nil
}

// implements gob encoding.
type GOBEncoder struct{}

func (enc GOBEncoder) Encode(w io.Writer, msg *RPC) error {
	return gob.NewEncoder(w).Encode(msg)
}
