// Package crypto implements encryption, decryption, and hash utilities for the DFS.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

// GenerateID generates a random 32-byte hexadecimal ID.
func GenerateID() string {
	buf := make([]byte, 32)
	io.ReadFull(rand.Reader, buf)
	return hex.EncodeToString(buf)
}

// HashKey hashes a string key using MD5 and returns its hexadecimal representation.
func HashKey(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

// NewEncryptionKey generates a random 32-byte key suitable for AES-256 encryption.
func NewEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuf)
	return keyBuf
}

// CopyStream reads from src, encrypts/decrypts using the given stream, and writes to dst.
// It returns the total number of bytes written to dst.
func CopyStream(stream cipher.Stream, blockSize int, src io.Reader, dst io.Writer) (int, error) {
	var (
		buf = make([]byte, 32*1024)
		nw  = blockSize
	)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return 0, nil
		}
	}
	return nw, nil
}

// CopyDecrypt reads the initialization vector (IV) from src, sets up an AES-CTR cipher stream with the given key,
// decrypts the rest of src, and writes it to dst.
func CopyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(block, iv)
	return CopyStream(stream, block.BlockSize(), src, dst)
}

// CopyEncrypt generates a random IV, writes it to dst, sets up an AES-CTR cipher stream with the given key,
// encrypts all data read from src, and writes it to dst.
func CopyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(block, iv)
	return CopyStream(stream, block.BlockSize(), src, dst)
}
