package msf

import (
	"fmt"
	"io"
)

// Stream provides sequential reading across non-contiguous blocks.
// It implements io.Reader, io.Seeker, and io.ReaderAt interfaces.
type Stream struct {
	data       io.ReaderAt
	blocks     []uint32
	blockSize  uint32
	streamSize uint32

	// Current position for Read/Seek
	pos uint32
}

// NewStream creates a new Stream reader for the given blocks.
func NewStream(data io.ReaderAt, blocks []uint32, blockSize, streamSize uint32) *Stream {
	return &Stream{
		data:       data,
		blocks:     blocks,
		blockSize:  blockSize,
		streamSize: streamSize,
		pos:        0,
	}
}

// Read implements io.Reader. It reads across block boundaries transparently.
func (s *Stream) Read(p []byte) (n int, err error) {
	if s.pos >= s.streamSize {
		return 0, io.EOF
	}

	remaining := s.streamSize - s.pos
	if uint32(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err = s.ReadAt(p, int64(s.pos))
	s.pos += uint32(n)
	return n, err
}

// ReadAt implements io.ReaderAt. It reads data at the given offset,
// handling block boundaries transparently.
func (s *Stream) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, fmt.Errorf("msf: negative offset: %d", off)
	}

	if off >= int64(s.streamSize) {
		return 0, io.EOF
	}

	pos := uint32(off)
	totalRead := 0

	for len(p) > 0 && pos < s.streamSize {
		// Calculate which block and offset within block
		blockIndex := pos / s.blockSize
		blockOffset := pos % s.blockSize

		if int(blockIndex) >= len(s.blocks) {
			return totalRead, io.EOF
		}

		// Calculate file offset for this block
		fileOffset := int64(s.blocks[blockIndex])*int64(s.blockSize) + int64(blockOffset)

		// How much can we read from this block?
		blockRemaining := s.blockSize - blockOffset
		streamRemaining := s.streamSize - pos
		toRead := uint32(len(p))

		if toRead > blockRemaining {
			toRead = blockRemaining
		}
		if toRead > streamRemaining {
			toRead = streamRemaining
		}

		// Read from the underlying data
		bytesRead, err := s.data.ReadAt(p[:toRead], fileOffset)
		totalRead += bytesRead
		p = p[bytesRead:]
		pos += uint32(bytesRead)

		if err != nil {
			if err == io.EOF && totalRead > 0 {
				// Partial read at end of file
				break
			}
			return totalRead, err
		}
	}

	if totalRead == 0 && int64(s.pos) >= int64(s.streamSize) {
		return 0, io.EOF
	}

	return totalRead, nil
}

// Seek implements io.Seeker.
func (s *Stream) Seek(offset int64, whence int) (int64, error) {
	var newPos int64

	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = int64(s.pos) + offset
	case io.SeekEnd:
		newPos = int64(s.streamSize) + offset
	default:
		return 0, fmt.Errorf("msf: invalid seek whence: %d", whence)
	}

	if newPos < 0 {
		return 0, fmt.Errorf("msf: negative seek position: %d", newPos)
	}

	if newPos > int64(s.streamSize) {
		newPos = int64(s.streamSize)
	}

	s.pos = uint32(newPos)
	return newPos, nil
}

// Size returns the total size of the stream in bytes.
func (s *Stream) Size() uint32 {
	return s.streamSize
}

// Position returns the current read position.
func (s *Stream) Position() uint32 {
	return s.pos
}

// Remaining returns the number of bytes remaining to be read.
func (s *Stream) Remaining() uint32 {
	if s.pos >= s.streamSize {
		return 0
	}
	return s.streamSize - s.pos
}

// Bytes reads the entire stream into a byte slice.
// This is useful for smaller streams that fit in memory.
func (s *Stream) Bytes() ([]byte, error) {
	data := make([]byte, s.streamSize)
	n, err := s.ReadAt(data, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return data[:n], nil
}

// Reset resets the stream position to the beginning.
func (s *Stream) Reset() {
	s.pos = 0
}
