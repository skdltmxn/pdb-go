// Package pdb provides parsing and querying of Microsoft PDB files.
package pdb

import (
	"errors"
	"fmt"
)

// Sentinel errors for common conditions.
var (
	// ErrNotPDB indicates the file is not a valid PDB.
	ErrNotPDB = errors.New("pdb: not a valid PDB file")

	// ErrUnsupportedVersion indicates an unsupported PDB version.
	ErrUnsupportedVersion = errors.New("pdb: unsupported PDB version")

	// ErrInvalidStream indicates a corrupted or invalid stream.
	ErrInvalidStream = errors.New("pdb: invalid stream")

	// ErrTypeNotFound indicates a type index was not found.
	ErrTypeNotFound = errors.New("pdb: type not found")

	// ErrSymbolNotFound indicates a symbol was not found.
	ErrSymbolNotFound = errors.New("pdb: symbol not found")

	// ErrModuleNotFound indicates a module was not found.
	ErrModuleNotFound = errors.New("pdb: module not found")

	// ErrFileClosed indicates the PDB file has been closed.
	ErrFileClosed = errors.New("pdb: file is closed")
)

// ParseError provides detailed information about parsing failures.
type ParseError struct {
	Stream  string // Stream name where error occurred
	Offset  int64  // Byte offset within stream
	Message string // Description of the error
	Err     error  // Underlying error, if any
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("pdb: parse error in %s at offset 0x%x: %s: %v",
			e.Stream, e.Offset, e.Message, e.Err)
	}
	return fmt.Sprintf("pdb: parse error in %s at offset 0x%x: %s",
		e.Stream, e.Offset, e.Message)
}

func (e *ParseError) Unwrap() error { return e.Err }
