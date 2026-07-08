// Package store implements the Content-Addressable Storage (CAS) layer on disk.
package store

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/keshavbansal015/DFS/crypto"
)

// DefaultRootFolderName is the fallback folder name used for storage if none is specified.
const DefaultRootFolderName = "storage_dir"

// CASPathTransformFunc implements a Content-Addressable Storage (CAS) path transformation.
// It hashes the key using SHA1, splits the hex string into 5-character blocks to form subdirectories,
// and returns a PathKey representing the directory pathname and filename.
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blocksize := 5
	sliceLen := len(hashStr) / blocksize

	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}
	return PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

// PathTransformFunc is a function signature that takes a key string and maps it to a PathKey.
type PathTransformFunc func(string) PathKey

// PathKey holds the pathname (representing the folder hierarchy) and filename (representing the file name).
type PathKey struct {
	// Pathname is the directory structure representation.
	Pathname string
	// Filename is the specific file name within that pathname.
	Filename string
}

// FirstPathName returns the top-level directory in the PathKey's pathname hierarchy.
func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.Pathname, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

// FullPath joins the Pathname and Filename to construct the relative file path.
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

// StoreOpts contains configuration options for the local Store.
type StoreOpts struct {
	// Root is the folder name of the root directory containing all files.
	Root string
	// PathTransformFunc is the mapping function used to translate keys to pathnames and filenames.
	PathTransformFunc PathTransformFunc
}

// DefaultPathTransformFunc is the fallback mapping function that returns the key unmodified as pathname and filename.
var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		Pathname: key,
		Filename: key,
	}
}

// Store represents the local physical storage on disk.
type Store struct {
	StoreOpts
}

// NewStore initializes a Store with the provided configuration options, applying defaults where needed.
func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = DefaultRootFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}

// Has checks if a file with the given key exists in the specified user/node ID directory in the storage root.
func (s *Store) Has(id, key string) bool {
	pathKey := s.PathTransformFunc(key)

	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)

	return !errors.Is(err, os.ErrNotExist)
}

// Clear recursively deletes the entire storage root directory.
func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

// Delete removes the file associated with the given key, alongside its parent directories,
// under the specified user/node ID folder.
func (s *Store) Delete(id, key string) error {
	pathKey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.Filename)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FirstPathName())
	return os.RemoveAll(firstPathNameWithRoot)
}

// Write writes the stream from the reader into a file identified by the key under the user/node ID.
// It returns the number of bytes written.
func (s *Store) Write(id string, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

// WriteDecrypt writes and decrypts (using the provided symmetric key) the stream from the reader into a file.
// It returns the number of bytes written.
func (s *Store) WriteDecrypt(encKey []byte, id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)

	if err != nil {
		return 0, err
	}
	n, err := crypto.CopyDecrypt(encKey, r, f)
	return int64(n), err
}

// openFileForWriting resolves the directory path, ensures it exists on disk, and creates or truncates the file.
func (s *Store) openFileForWriting(id string, key string) (*os.File, error) {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.Pathname)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	return os.Create(fullPathWithRoot)
}

// writeStream copies the content from the reader directly to the target file.
func (s *Store) writeStream(id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)

	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}

// Read opens the requested file and returns its size and an io.Reader.
func (s *Store) Read(id, key string) (int64, io.Reader, error) {
	return s.readStream(id, key)
}

// readStream resolves the full path, opens the file, stats it to get its size, and returns it.
func (s *Store) readStream(id, key string) (int64, io.Reader, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	file, err := os.Open(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil
}
