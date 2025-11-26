package msf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

// File represents an opened MSF file.
// It provides access to the underlying streams in a thread-safe manner.
type File struct {
	data      io.ReaderAt
	closer    io.Closer // may be nil if data doesn't need closing
	size      int64
	superBlock *SuperBlock
	directory  *StreamDirectory

	// Lazy loading synchronization
	dirOnce sync.Once
	dirErr  error

	mu sync.RWMutex
}

// Open opens an MSF file from the given path.
func Open(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("msf: failed to open file: %w", err)
	}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("msf: failed to stat file: %w", err)
	}

	msf, err := NewFile(f, stat.Size())
	if err != nil {
		f.Close()
		return nil, err
	}

	msf.closer = f
	return msf, nil
}

// NewFile creates an MSF file from an io.ReaderAt.
// This allows reading from arbitrary sources (embedded, network, etc.)
// The caller is responsible for closing the underlying reader if needed.
func NewFile(r io.ReaderAt, size int64) (*File, error) {
	if size < SuperBlockSize {
		return nil, ErrTruncatedFile
	}

	// Read and validate superblock
	sbData := make([]byte, SuperBlockSize)
	if _, err := r.ReadAt(sbData, 0); err != nil {
		return nil, fmt.Errorf("msf: failed to read superblock: %w", err)
	}

	sb, err := ReadSuperBlock(bytes.NewReader(sbData))
	if err != nil {
		return nil, err
	}

	// Validate file size matches expected size
	expectedSize := sb.FileSize()
	if size < expectedSize {
		return nil, fmt.Errorf("msf: file too small: got %d bytes, expected %d", size, expectedSize)
	}

	return &File{
		data:       r,
		size:       size,
		superBlock: sb,
	}, nil
}

// Close releases resources associated with the MSF file.
func (f *File) Close() error {
	if f.closer != nil {
		return f.closer.Close()
	}
	return nil
}

// SuperBlock returns the MSF superblock.
func (f *File) SuperBlock() *SuperBlock {
	return f.superBlock
}

// Directory returns the stream directory.
// The directory is lazily loaded on first access.
func (f *File) Directory() (*StreamDirectory, error) {
	f.dirOnce.Do(func() {
		dr := NewDirectoryReader(f.superBlock, f.data)
		f.directory, f.dirErr = dr.ReadDirectory()
	})

	if f.dirErr != nil {
		return nil, f.dirErr
	}
	return f.directory, nil
}

// NumStreams returns the number of streams in the file.
func (f *File) NumStreams() (uint32, error) {
	dir, err := f.Directory()
	if err != nil {
		return 0, err
	}
	return dir.NumStreams, nil
}

// StreamSize returns the size of the given stream in bytes.
func (f *File) StreamSize(streamIndex uint32) (uint32, error) {
	dir, err := f.Directory()
	if err != nil {
		return 0, err
	}
	return dir.StreamSize(streamIndex), nil
}

// StreamExists returns true if the stream exists and is not a nil stream.
func (f *File) StreamExists(streamIndex uint32) (bool, error) {
	dir, err := f.Directory()
	if err != nil {
		return false, err
	}
	return dir.StreamExists(streamIndex), nil
}

// OpenStream opens a stream for reading.
// Returns an error if the stream doesn't exist.
func (f *File) OpenStream(streamIndex uint32) (*Stream, error) {
	dir, err := f.Directory()
	if err != nil {
		return nil, err
	}

	if streamIndex >= dir.NumStreams {
		return nil, fmt.Errorf("%w: %d", ErrInvalidStreamIndex, streamIndex)
	}

	size := dir.StreamSizes[streamIndex]
	if size == NilStreamSize {
		return nil, fmt.Errorf("msf: stream %d is nil", streamIndex)
	}

	blocks := dir.StreamBlocks[streamIndex]
	return NewStream(f.data, blocks, f.superBlock.BlockSize, size), nil
}

// ReadStream reads an entire stream into memory.
// This is a convenience method for smaller streams.
func (f *File) ReadStream(streamIndex uint32) ([]byte, error) {
	stream, err := f.OpenStream(streamIndex)
	if err != nil {
		return nil, err
	}
	return stream.Bytes()
}

// BlockSize returns the block size used by this MSF file.
func (f *File) BlockSize() uint32 {
	return f.superBlock.BlockSize
}

// FileSize returns the total size of the MSF file.
func (f *File) FileSize() int64 {
	return f.size
}

// NumBlocks returns the total number of blocks in the file.
func (f *File) NumBlocks() uint32 {
	return f.superBlock.NumBlocks
}
