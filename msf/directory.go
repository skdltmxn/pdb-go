package msf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// NilStreamSize indicates a deleted or nil stream
const NilStreamSize = 0xFFFFFFFF

// Well-known stream indices
const (
	StreamOldDirectory = 0 // Old MSF directory (unused in PDB 7.0)
	StreamPDBInfo      = 1 // PDB Info stream (GUID, age, named streams)
	StreamTPI          = 2 // Type Program Information
	StreamDBI          = 3 // Debug Information
	StreamIPI          = 4 // ID Program Information
)

// Directory parsing errors
var (
	ErrTruncatedDirectory   = errors.New("msf: truncated stream directory")
	ErrInvalidStreamIndex   = errors.New("msf: invalid stream index")
	ErrInvalidBlockIndex    = errors.New("msf: invalid block index")
	ErrDirectoryBlockMapNil = errors.New("msf: directory block map is nil")
)

// StreamDirectory describes all streams in the MSF file.
// It is a jagged array where each stream has its own list of block indices.
type StreamDirectory struct {
	// NumStreams is the count of streams
	NumStreams uint32

	// StreamSizes holds the size in bytes of each stream.
	// A value of NilStreamSize (0xFFFFFFFF) indicates a deleted stream.
	StreamSizes []uint32

	// StreamBlocks is a jagged array where StreamBlocks[i] contains
	// the block indices for stream i. For nil streams, this will be nil.
	StreamBlocks [][]uint32
}

// ParseDirectory reads the stream directory from the given byte slice.
// The data should be the concatenated contents of all directory blocks.
func ParseDirectory(data []byte, blockSize uint32) (*StreamDirectory, error) {
	if len(data) < 4 {
		return nil, ErrTruncatedDirectory
	}

	dir := &StreamDirectory{}
	offset := 0

	// Read number of streams
	dir.NumStreams = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// Validate we have enough space for stream sizes
	expectedSizeBytes := int(dir.NumStreams) * 4
	if len(data) < offset+expectedSizeBytes {
		return nil, ErrTruncatedDirectory
	}

	// Read stream sizes
	dir.StreamSizes = make([]uint32, dir.NumStreams)
	for i := uint32(0); i < dir.NumStreams; i++ {
		dir.StreamSizes[i] = binary.LittleEndian.Uint32(data[offset:])
		offset += 4
	}

	// Read block indices for each stream
	dir.StreamBlocks = make([][]uint32, dir.NumStreams)
	for i := uint32(0); i < dir.NumStreams; i++ {
		size := dir.StreamSizes[i]
		if size == NilStreamSize || size == 0 {
			// Nil or empty stream has no blocks
			continue
		}

		// Calculate number of blocks for this stream
		numBlocks := (size + blockSize - 1) / blockSize
		dir.StreamBlocks[i] = make([]uint32, numBlocks)

		for j := uint32(0); j < numBlocks; j++ {
			if offset+4 > len(data) {
				return nil, ErrTruncatedDirectory
			}
			dir.StreamBlocks[i][j] = binary.LittleEndian.Uint32(data[offset:])
			offset += 4
		}
	}

	return dir, nil
}

// StreamSize returns the size of the given stream, or 0 if the stream doesn't exist
// or is a nil stream.
func (d *StreamDirectory) StreamSize(streamIndex uint32) uint32 {
	if streamIndex >= d.NumStreams {
		return 0
	}
	size := d.StreamSizes[streamIndex]
	if size == NilStreamSize {
		return 0
	}
	return size
}

// StreamExists returns true if the stream exists and is not a nil stream.
func (d *StreamDirectory) StreamExists(streamIndex uint32) bool {
	if streamIndex >= d.NumStreams {
		return false
	}
	return d.StreamSizes[streamIndex] != NilStreamSize && d.StreamSizes[streamIndex] > 0
}

// GetStreamBlocks returns the block indices for the given stream.
// Returns an error if the stream doesn't exist.
func (d *StreamDirectory) GetStreamBlocks(streamIndex uint32) ([]uint32, error) {
	if streamIndex >= d.NumStreams {
		return nil, fmt.Errorf("%w: %d >= %d", ErrInvalidStreamIndex, streamIndex, d.NumStreams)
	}
	if d.StreamSizes[streamIndex] == NilStreamSize {
		return nil, nil
	}
	return d.StreamBlocks[streamIndex], nil
}

// DirectoryReader helps read the stream directory from an MSF file.
// It handles the indirection through the block map address.
type DirectoryReader struct {
	sb   *SuperBlock
	data io.ReaderAt
}

// NewDirectoryReader creates a new DirectoryReader.
func NewDirectoryReader(sb *SuperBlock, data io.ReaderAt) *DirectoryReader {
	return &DirectoryReader{sb: sb, data: data}
}

// ReadDirectory reads and parses the complete stream directory.
func (dr *DirectoryReader) ReadDirectory() (*StreamDirectory, error) {
	// First, read the block map (array of block indices for the directory)
	blockMapBytes, err := dr.readBlockMap()
	if err != nil {
		return nil, err
	}

	// Now read all directory blocks and concatenate them
	directoryData, err := dr.readDirectoryBlocks(blockMapBytes)
	if err != nil {
		return nil, err
	}

	// Parse the directory
	return ParseDirectory(directoryData, dr.sb.BlockSize)
}

// readBlockMap reads the array of block indices that make up the stream directory.
func (dr *DirectoryReader) readBlockMap() ([]uint32, error) {
	numDirectoryBlocks := dr.sb.NumDirectoryBlocks()

	// Calculate how many bytes the block map takes
	blockMapSize := numDirectoryBlocks * 4

	// Calculate how many blocks the block map spans
	numBlockMapBlocks := (blockMapSize + dr.sb.BlockSize - 1) / dr.sb.BlockSize

	// Read block indices from BlockMapAddr
	blockMapData := make([]byte, numBlockMapBlocks*dr.sb.BlockSize)
	for i := uint32(0); i < numBlockMapBlocks; i++ {
		blockOffset := dr.sb.BlockOffset(dr.sb.BlockMapAddr + i)
		_, err := dr.data.ReadAt(blockMapData[i*dr.sb.BlockSize:(i+1)*dr.sb.BlockSize], blockOffset)
		if err != nil {
			return nil, fmt.Errorf("msf: failed to read block map: %w", err)
		}
	}

	// Parse block indices
	blockMap := make([]uint32, numDirectoryBlocks)
	for i := uint32(0); i < numDirectoryBlocks; i++ {
		blockMap[i] = binary.LittleEndian.Uint32(blockMapData[i*4:])
	}

	return blockMap, nil
}

// readDirectoryBlocks reads and concatenates all directory blocks.
func (dr *DirectoryReader) readDirectoryBlocks(blockIndices []uint32) ([]byte, error) {
	directoryData := make([]byte, dr.sb.NumDirectoryBytes)
	remaining := dr.sb.NumDirectoryBytes

	for i, blockIdx := range blockIndices {
		if blockIdx >= dr.sb.NumBlocks {
			return nil, fmt.Errorf("%w: %d >= %d", ErrInvalidBlockIndex, blockIdx, dr.sb.NumBlocks)
		}

		toRead := dr.sb.BlockSize
		if toRead > remaining {
			toRead = remaining
		}

		blockOffset := dr.sb.BlockOffset(blockIdx)
		destOffset := uint32(i) * dr.sb.BlockSize
		_, err := dr.data.ReadAt(directoryData[destOffset:destOffset+toRead], blockOffset)
		if err != nil {
			return nil, fmt.Errorf("msf: failed to read directory block %d: %w", blockIdx, err)
		}

		remaining -= toRead
	}

	return directoryData, nil
}
