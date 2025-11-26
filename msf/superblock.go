// Package msf provides parsing for the MSF (Multi-Stream File) container format
// used by Microsoft PDB files.
package msf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Magic signature for PDB 7.0 format (BigMsf)
const Magic = "Microsoft C/C++ MSF 7.00\r\n\x1a\x44\x53\x00\x00\x00"

// MagicSize is the size of the magic signature in bytes
const MagicSize = 32

// SuperBlockSize is the total size of the SuperBlock structure
const SuperBlockSize = 56

// Valid block sizes for MSF files
const (
	BlockSizeMin   uint32 = 512   // Minimum block size
	BlockSize512   uint32 = 512
	BlockSize1024  uint32 = 1024
	BlockSize2048  uint32 = 2048
	BlockSize4096  uint32 = 4096  // "BigMsf" - most common
	BlockSize8192  uint32 = 8192
	BlockSize16384 uint32 = 16384
	BlockSize32768 uint32 = 32768
	BlockSizeMax   uint32 = 65536 // Maximum block size
)

// Errors returned during SuperBlock parsing
var (
	ErrInvalidMagic     = errors.New("msf: invalid magic signature, not a valid PDB file")
	ErrInvalidBlockSize = errors.New("msf: invalid block size")
	ErrInvalidFPMBlock  = errors.New("msf: invalid free block map block index")
	ErrTruncatedFile    = errors.New("msf: file is truncated")
)

// SuperBlock is located at file offset 0 and describes the MSF container structure.
// It contains metadata about the file's block-based layout and the location of
// the stream directory.
type SuperBlock struct {
	// FileMagic must equal the Magic constant (32 bytes)
	FileMagic [MagicSize]byte

	// BlockSize is the internal file system block size (512, 1024, 2048, or 4096)
	BlockSize uint32

	// FreeBlockMapBlock is the index of the active FPM block (always 1 or 2)
	// The MSF format supports atomic updates by writing to the inactive FPM,
	// then swapping this value.
	FreeBlockMapBlock uint32

	// NumBlocks is the total number of blocks in the file.
	// NumBlocks * BlockSize should equal the file size.
	NumBlocks uint32

	// NumDirectoryBytes is the size of the stream directory in bytes
	NumDirectoryBytes uint32

	// Unknown is a reserved field (always 0)
	Unknown uint32

	// BlockMapAddr is the block index containing the array of block indices
	// that make up the stream directory. For large directories spanning multiple
	// blocks, this provides an extra level of indirection.
	BlockMapAddr uint32
}

// ReadSuperBlock reads and validates a SuperBlock from the given reader.
// The reader should be positioned at the beginning of the PDB file.
func ReadSuperBlock(r io.Reader) (*SuperBlock, error) {
	var sb SuperBlock

	if err := binary.Read(r, binary.LittleEndian, &sb); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, ErrTruncatedFile
		}
		return nil, fmt.Errorf("msf: failed to read superblock: %w", err)
	}

	if err := sb.Validate(); err != nil {
		return nil, err
	}

	return &sb, nil
}

// Validate checks the SuperBlock for internal consistency.
func (sb *SuperBlock) Validate() error {
	// Check magic signature
	if string(sb.FileMagic[:]) != Magic {
		return ErrInvalidMagic
	}

	// Check block size is a power of 2 within valid range
	if sb.BlockSize < BlockSizeMin || sb.BlockSize > BlockSizeMax {
		return ErrInvalidBlockSize
	}
	// Block size must be a power of 2
	if sb.BlockSize&(sb.BlockSize-1) != 0 {
		return ErrInvalidBlockSize
	}

	// FreeBlockMapBlock must be 1 or 2
	if sb.FreeBlockMapBlock != 1 && sb.FreeBlockMapBlock != 2 {
		return ErrInvalidFPMBlock
	}

	return nil
}

// NumDirectoryBlocks returns the number of blocks needed to store the stream directory.
func (sb *SuperBlock) NumDirectoryBlocks() uint32 {
	return (sb.NumDirectoryBytes + sb.BlockSize - 1) / sb.BlockSize
}

// FileSize returns the expected file size based on NumBlocks and BlockSize.
func (sb *SuperBlock) FileSize() int64 {
	return int64(sb.NumBlocks) * int64(sb.BlockSize)
}

// BlockOffset returns the byte offset of the given block number.
func (sb *SuperBlock) BlockOffset(blockNum uint32) int64 {
	return int64(blockNum) * int64(sb.BlockSize)
}
