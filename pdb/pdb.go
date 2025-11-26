package pdb

import (
	"fmt"
	"io"
	"sync"

	"github.com/skdltmxn/pdb-go/internal/dbi"
	"github.com/skdltmxn/pdb-go/internal/tpi"
	"github.com/skdltmxn/pdb-go/msf"
)

// File represents an opened PDB file.
// It is safe for concurrent read access after opening.
type File struct {
	msf    *msf.File
	closed bool
	mu     sync.RWMutex

	// Lazy-loaded streams
	pdbInfo     *PDBInfo
	pdbInfoOnce sync.Once
	pdbInfoErr  error

	tpiStream     *tpi.Stream
	tpiStreamOnce sync.Once
	tpiStreamErr  error

	ipiStream     *tpi.Stream
	ipiStreamOnce sync.Once
	ipiStreamErr  error

	dbiStream     *dbi.Stream
	dbiStreamOnce sync.Once
	dbiStreamErr  error

	// Cached data
	symbolTable     *SymbolTable
	symbolTableOnce sync.Once
	symbolTableErr  error

	typeTable     *TypeTable
	typeTableOnce sync.Once
	typeTableErr  error
}

// PDBInfo contains metadata about the PDB file.
type PDBInfo struct {
	Version   uint32
	Signature uint32
	Age       uint32
	GUID      [16]byte
}

// Open opens a PDB file from the given path.
func Open(path string) (*File, error) {
	msfFile, err := msf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("pdb: failed to open file: %w", err)
	}

	return &File{msf: msfFile}, nil
}

// OpenReader opens a PDB from an io.ReaderAt.
// This allows reading from arbitrary sources (embedded, network, etc.)
func OpenReader(r io.ReaderAt, size int64) (*File, error) {
	msfFile, err := msf.NewFile(r, size)
	if err != nil {
		return nil, fmt.Errorf("pdb: failed to open file: %w", err)
	}

	return &File{msf: msfFile}, nil
}

// Close releases resources associated with the PDB file.
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true
	return f.msf.Close()
}

// Info returns metadata about the PDB file.
func (f *File) Info() (*PDBInfo, error) {
	f.pdbInfoOnce.Do(func() {
		f.pdbInfo, f.pdbInfoErr = f.loadPDBInfo()
	})

	if f.pdbInfoErr != nil {
		return nil, f.pdbInfoErr
	}
	return f.pdbInfo, nil
}

func (f *File) loadPDBInfo() (*PDBInfo, error) {
	data, err := f.msf.ReadStream(msf.StreamPDBInfo)
	if err != nil {
		return nil, fmt.Errorf("pdb: failed to read PDB info stream: %w", err)
	}

	if len(data) < 28 {
		return nil, fmt.Errorf("pdb: PDB info stream too short")
	}

	info := &PDBInfo{}
	info.Version = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	info.Signature = uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24
	info.Age = uint32(data[8]) | uint32(data[9])<<8 | uint32(data[10])<<16 | uint32(data[11])<<24
	copy(info.GUID[:], data[12:28])

	return info, nil
}

// Symbols returns a symbol table for querying symbols.
func (f *File) Symbols() (*SymbolTable, error) {
	f.symbolTableOnce.Do(func() {
		f.symbolTable, f.symbolTableErr = f.loadSymbolTable()
	})

	if f.symbolTableErr != nil {
		return nil, f.symbolTableErr
	}
	return f.symbolTable, nil
}

func (f *File) loadSymbolTable() (*SymbolTable, error) {
	dbiStream, err := f.getDBI()
	if err != nil {
		return nil, err
	}

	st := newSymbolTable(f, dbiStream)
	return st, nil
}

// Types returns a type table for querying type information.
func (f *File) Types() (*TypeTable, error) {
	f.typeTableOnce.Do(func() {
		f.typeTable, f.typeTableErr = f.loadTypeTable()
	})

	if f.typeTableErr != nil {
		return nil, f.typeTableErr
	}
	return f.typeTable, nil
}

func (f *File) loadTypeTable() (*TypeTable, error) {
	tpiStream, err := f.getTPI()
	if err != nil {
		return nil, err
	}

	return newTypeTable(tpiStream), nil
}

// Modules returns all modules (compilands) in the PDB.
func (f *File) Modules() ([]*Module, error) {
	dbiStream, err := f.getDBI()
	if err != nil {
		return nil, err
	}

	modules := make([]*Module, len(dbiStream.Modules))
	for i := range dbiStream.Modules {
		modules[i] = &Module{
			pdb:   f,
			index: i,
			info:  &dbiStream.Modules[i],
		}
	}

	return modules, nil
}

// ModuleCount returns the number of modules in the PDB.
func (f *File) ModuleCount() (int, error) {
	dbiStream, err := f.getDBI()
	if err != nil {
		return 0, err
	}
	return len(dbiStream.Modules), nil
}

// BlockSize returns the block size used by this PDB file.
func (f *File) BlockSize() uint32 {
	return f.msf.BlockSize()
}

// NumStreams returns the number of streams in the PDB.
func (f *File) NumStreams() (uint32, error) {
	return f.msf.NumStreams()
}

// Internal helpers

func (f *File) getTPI() (*tpi.Stream, error) {
	f.tpiStreamOnce.Do(func() {
		data, err := f.msf.ReadStream(msf.StreamTPI)
		if err != nil {
			f.tpiStreamErr = fmt.Errorf("pdb: failed to read TPI stream: %w", err)
			return
		}

		f.tpiStream, f.tpiStreamErr = tpi.ParseStream(data)
	})

	if f.tpiStreamErr != nil {
		return nil, f.tpiStreamErr
	}
	return f.tpiStream, nil
}

func (f *File) getIPI() (*tpi.Stream, error) {
	f.ipiStreamOnce.Do(func() {
		exists, err := f.msf.StreamExists(msf.StreamIPI)
		if err != nil || !exists {
			f.ipiStreamErr = fmt.Errorf("pdb: IPI stream not found")
			return
		}

		data, err := f.msf.ReadStream(msf.StreamIPI)
		if err != nil {
			f.ipiStreamErr = fmt.Errorf("pdb: failed to read IPI stream: %w", err)
			return
		}

		f.ipiStream, f.ipiStreamErr = tpi.ParseStream(data)
	})

	if f.ipiStreamErr != nil {
		return nil, f.ipiStreamErr
	}
	return f.ipiStream, nil
}

func (f *File) getDBI() (*dbi.Stream, error) {
	f.dbiStreamOnce.Do(func() {
		data, err := f.msf.ReadStream(msf.StreamDBI)
		if err != nil {
			f.dbiStreamErr = fmt.Errorf("pdb: failed to read DBI stream: %w", err)
			return
		}

		f.dbiStream, f.dbiStreamErr = dbi.ParseStream(data)
	})

	if f.dbiStreamErr != nil {
		return nil, f.dbiStreamErr
	}
	return f.dbiStream, nil
}

func (f *File) readModuleSymbols(streamIndex uint16) ([]byte, error) {
	if streamIndex == 0xFFFF {
		return nil, nil
	}

	return f.msf.ReadStream(uint32(streamIndex))
}
