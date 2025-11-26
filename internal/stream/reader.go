// Package stream provides binary reading utilities for PDB parsing.
package stream

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// Errors returned by Reader
var (
	ErrUnexpectedEOF   = errors.New("stream: unexpected end of data")
	ErrInvalidString   = errors.New("stream: invalid string encoding")
	ErrNegativeOffset  = errors.New("stream: negative offset")
	ErrInvalidNumeric  = errors.New("stream: invalid numeric encoding")
)

// Reader provides methods for reading binary data from PDB streams.
// All multi-byte values are read in little-endian order.
type Reader struct {
	data   []byte
	offset int
}

// NewReader creates a Reader from a byte slice.
func NewReader(data []byte) *Reader {
	return &Reader{data: data, offset: 0}
}

// Offset returns the current read position.
func (r *Reader) Offset() int {
	return r.offset
}

// SetOffset sets the read position.
func (r *Reader) SetOffset(offset int) error {
	if offset < 0 {
		return ErrNegativeOffset
	}
	r.offset = offset
	return nil
}

// Remaining returns the number of bytes remaining.
func (r *Reader) Remaining() int {
	if r.offset >= len(r.data) {
		return 0
	}
	return len(r.data) - r.offset
}

// Skip advances the read position by n bytes.
func (r *Reader) Skip(n int) error {
	if r.offset+n > len(r.data) {
		return ErrUnexpectedEOF
	}
	r.offset += n
	return nil
}

// Align aligns the read position to the given boundary.
func (r *Reader) Align(alignment int) {
	if alignment <= 1 {
		return
	}
	mod := r.offset % alignment
	if mod != 0 {
		r.offset += alignment - mod
	}
}

// ReadU8 reads an unsigned 8-bit integer.
func (r *Reader) ReadU8() (uint8, error) {
	if r.offset >= len(r.data) {
		return 0, ErrUnexpectedEOF
	}
	v := r.data[r.offset]
	r.offset++
	return v, nil
}

// ReadU16 reads an unsigned 16-bit integer.
func (r *Reader) ReadU16() (uint16, error) {
	if r.offset+2 > len(r.data) {
		return 0, ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint16(r.data[r.offset:])
	r.offset += 2
	return v, nil
}

// ReadU32 reads an unsigned 32-bit integer.
func (r *Reader) ReadU32() (uint32, error) {
	if r.offset+4 > len(r.data) {
		return 0, ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(r.data[r.offset:])
	r.offset += 4
	return v, nil
}

// ReadU64 reads an unsigned 64-bit integer.
func (r *Reader) ReadU64() (uint64, error) {
	if r.offset+8 > len(r.data) {
		return 0, ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint64(r.data[r.offset:])
	r.offset += 8
	return v, nil
}

// ReadI8 reads a signed 8-bit integer.
func (r *Reader) ReadI8() (int8, error) {
	v, err := r.ReadU8()
	return int8(v), err
}

// ReadI16 reads a signed 16-bit integer.
func (r *Reader) ReadI16() (int16, error) {
	v, err := r.ReadU16()
	return int16(v), err
}

// ReadI32 reads a signed 32-bit integer.
func (r *Reader) ReadI32() (int32, error) {
	v, err := r.ReadU32()
	return int32(v), err
}

// ReadI64 reads a signed 64-bit integer.
func (r *Reader) ReadI64() (int64, error) {
	v, err := r.ReadU64()
	return int64(v), err
}

// ReadFloat32 reads a 32-bit float.
func (r *Reader) ReadFloat32() (float32, error) {
	v, err := r.ReadU32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(v), nil
}

// ReadFloat64 reads a 64-bit float.
func (r *Reader) ReadFloat64() (float64, error) {
	v, err := r.ReadU64()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(v), nil
}

// ReadBytes reads n bytes.
func (r *Reader) ReadBytes(n int) ([]byte, error) {
	if r.offset+n > len(r.data) {
		return nil, ErrUnexpectedEOF
	}
	v := make([]byte, n)
	copy(v, r.data[r.offset:r.offset+n])
	r.offset += n
	return v, nil
}

// ReadBytesRef returns a reference to n bytes without copying.
// The returned slice is only valid as long as the underlying data.
func (r *Reader) ReadBytesRef(n int) ([]byte, error) {
	if r.offset+n > len(r.data) {
		return nil, ErrUnexpectedEOF
	}
	v := r.data[r.offset : r.offset+n]
	r.offset += n
	return v, nil
}

// ReadCString reads a null-terminated string.
func (r *Reader) ReadCString() (string, error) {
	start := r.offset
	for r.offset < len(r.data) {
		if r.data[r.offset] == 0 {
			s := string(r.data[start:r.offset])
			r.offset++ // Skip null terminator
			return s, nil
		}
		r.offset++
	}
	return "", ErrUnexpectedEOF
}

// ReadFixedString reads a fixed-length string, trimming any null padding.
func (r *Reader) ReadFixedString(n int) (string, error) {
	if r.offset+n > len(r.data) {
		return "", ErrUnexpectedEOF
	}
	data := r.data[r.offset : r.offset+n]
	r.offset += n

	// Trim null bytes from the end
	end := len(data)
	for end > 0 && data[end-1] == 0 {
		end--
	}
	return string(data[:end]), nil
}

// ReadGUID reads a 16-byte GUID.
func (r *Reader) ReadGUID() ([16]byte, error) {
	var guid [16]byte
	if r.offset+16 > len(r.data) {
		return guid, ErrUnexpectedEOF
	}
	copy(guid[:], r.data[r.offset:r.offset+16])
	r.offset += 16
	return guid, nil
}

// ReadNumeric reads a CodeView encoded numeric value.
// This handles the variable-length encoding used in type records.
func (r *Reader) ReadNumeric() (uint64, error) {
	leaf, err := r.ReadU16()
	if err != nil {
		return 0, err
	}

	// Values less than 0x8000 are the value itself
	if leaf < 0x8000 {
		return uint64(leaf), nil
	}

	// Otherwise, leaf indicates the type of the following value
	switch leaf {
	case 0x8000: // LF_CHAR
		v, err := r.ReadI8()
		return uint64(v), err
	case 0x8001: // LF_SHORT
		v, err := r.ReadI16()
		return uint64(v), err
	case 0x8002: // LF_USHORT
		v, err := r.ReadU16()
		return uint64(v), err
	case 0x8003: // LF_LONG
		v, err := r.ReadI32()
		return uint64(v), err
	case 0x8004: // LF_ULONG
		v, err := r.ReadU32()
		return uint64(v), err
	case 0x8009: // LF_QUADWORD
		v, err := r.ReadI64()
		return uint64(v), err
	case 0x800a: // LF_UQUADWORD
		return r.ReadU64()
	default:
		return 0, ErrInvalidNumeric
	}
}

// Peek returns a copy of the next n bytes without advancing the position.
func (r *Reader) Peek(n int) ([]byte, error) {
	if r.offset+n > len(r.data) {
		return nil, ErrUnexpectedEOF
	}
	v := make([]byte, n)
	copy(v, r.data[r.offset:r.offset+n])
	return v, nil
}

// PeekU8 returns the next byte without advancing the position.
func (r *Reader) PeekU8() (uint8, error) {
	if r.offset >= len(r.data) {
		return 0, ErrUnexpectedEOF
	}
	return r.data[r.offset], nil
}

// PeekU16 returns the next 16-bit integer without advancing the position.
func (r *Reader) PeekU16() (uint16, error) {
	if r.offset+2 > len(r.data) {
		return 0, ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint16(r.data[r.offset:]), nil
}

// Read implements io.Reader.
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

// Slice returns a new Reader for a subset of the data.
func (r *Reader) Slice(offset, length int) (*Reader, error) {
	if offset < 0 || offset+length > len(r.data) {
		return nil, ErrUnexpectedEOF
	}
	return NewReader(r.data[offset : offset+length]), nil
}

// SubReader returns a new Reader starting at the current position with the given length.
func (r *Reader) SubReader(length int) (*Reader, error) {
	if r.offset+length > len(r.data) {
		return nil, ErrUnexpectedEOF
	}
	sub := NewReader(r.data[r.offset : r.offset+length])
	r.offset += length
	return sub, nil
}

// Data returns the underlying byte slice.
func (r *Reader) Data() []byte {
	return r.data
}

// RemainingData returns the remaining unread data.
func (r *Reader) RemainingData() []byte {
	if r.offset >= len(r.data) {
		return nil
	}
	return r.data[r.offset:]
}
