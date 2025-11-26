package pdb

import (
	"iter"
	"sync"

	"github.com/skdltmxn/pdb-go/internal/dbi"
	"github.com/skdltmxn/pdb-go/internal/symbols"
)

// Module represents a compilation unit (object file) in the PDB.
type Module struct {
	pdb   *File
	index int
	info  *dbi.ModuleInfo

	// Lazy-loaded symbols
	symbols     []Symbol
	symbolsOnce sync.Once
	symbolsErr  error
}

// Index returns the module index.
func (m *Module) Index() int {
	return m.index
}

// Name returns the module name (typically the object file path).
func (m *Module) Name() string {
	return m.info.ModuleName
}

// ObjectFileName returns the original object file name.
func (m *Module) ObjectFileName() string {
	return m.info.ObjFileName
}

// Section returns the section index for this module's contribution.
func (m *Module) Section() uint16 {
	return m.info.Section.Section
}

// Offset returns the offset within the section.
func (m *Module) Offset() int32 {
	return m.info.Section.Offset
}

// Size returns the size of this module's contribution.
func (m *Module) Size() int32 {
	return m.info.Section.Size
}

// SourceFileCount returns the number of source files.
func (m *Module) SourceFileCount() uint16 {
	return m.info.SourceFileCount
}

// Symbols returns an iterator over symbols in this module.
func (m *Module) Symbols() iter.Seq[Symbol] {
	return func(yield func(Symbol) bool) {
		m.loadSymbols()

		if m.symbolsErr != nil {
			return
		}

		for _, sym := range m.symbols {
			if !yield(sym) {
				return
			}
		}
	}
}

func (m *Module) loadSymbols() {
	m.symbolsOnce.Do(func() {
		m.symbols, m.symbolsErr = m.parseSymbols()
	})
}

func (m *Module) parseSymbols() ([]Symbol, error) {
	// Get module symbol stream data
	data, err := m.pdb.readModuleSymbols(m.info.ModuleSymStreamIndex)
	if err != nil {
		return nil, err
	}
	if data == nil || len(data) == 0 {
		return nil, nil
	}

	// The module stream starts with a signature, then symbol records
	if len(data) < 4 {
		return nil, nil
	}

	// Skip signature (4 bytes)
	symData := data[4:]
	if uint32(len(symData)) < m.info.SymByteSize-4 {
		symData = symData[:m.info.SymByteSize-4]
	}

	// Parse symbol records
	iter := symbols.NewSymbolIterator(symData)
	var result []Symbol

	for {
		record, err := iter.Next()
		if err != nil {
			break
		}
		if record == nil {
			break
		}

		sym := m.convertSymbol(record)
		if sym != nil {
			result = append(result, sym)
		}
	}

	return result, nil
}

func (m *Module) convertSymbol(record *symbols.SymbolRecord) Symbol {
	switch record.Kind {
	case symbols.S_GPROC32, symbols.S_LPROC32, symbols.S_GPROC32_ID, symbols.S_LPROC32_ID:
		proc, err := symbols.ParseProcSym(record.Data)
		if err != nil {
			return nil
		}
		return &FunctionSymbol{
			baseSymbol: baseSymbol{name: proc.Name},
			section:    proc.Segment,
			offset:     proc.CodeOffset,
			length:     proc.CodeSize,
			typeIndex:  uint32(proc.FunctionType),
		}

	case symbols.S_GDATA32, symbols.S_LDATA32:
		data, err := symbols.ParseDataSym(record.Data)
		if err != nil {
			return nil
		}
		return &DataSymbol{
			baseSymbol: baseSymbol{name: data.Name},
			section:    data.Segment,
			offset:     data.Offset,
			typeIndex:  uint32(data.Type),
		}

	case symbols.S_UDT, symbols.S_UDT_ST:
		udt, err := symbols.ParseUDTSym(record.Data)
		if err != nil {
			return nil
		}
		return &UDTSymbol{
			baseSymbol: baseSymbol{name: udt.Name},
			typeIndex:  uint32(udt.Type),
		}

	case symbols.S_CONSTANT, symbols.S_CONSTANT_ST:
		c, err := symbols.ParseConstantSym(record.Data)
		if err != nil {
			return nil
		}
		return &ConstantSymbol{
			baseSymbol: baseSymbol{name: c.Name},
			value:      c.Value,
			typeIndex:  uint32(c.Type),
		}

	case symbols.S_LOCAL:
		local, err := symbols.ParseLocalSym(record.Data)
		if err != nil {
			return nil
		}
		return &LocalSymbol{
			baseSymbol:  baseSymbol{name: local.Name},
			typeIndex:   uint32(local.Type),
			isParameter: local.Flags.IsParameter(),
		}

	case symbols.S_LABEL32:
		label, err := symbols.ParseLabelSym(record.Data)
		if err != nil {
			return nil
		}
		return &LabelSymbol{
			baseSymbol: baseSymbol{name: label.Name},
			section:    label.Segment,
			offset:     label.Offset,
		}

	case symbols.S_BLOCK32:
		block, err := symbols.ParseBlockSym(record.Data)
		if err != nil {
			return nil
		}
		return &BlockSymbol{
			baseSymbol: baseSymbol{name: block.Name},
			section:    block.Segment,
			offset:     block.Offset,
			length:     block.CodeSize,
		}

	case symbols.S_THUNK32:
		thunk, err := symbols.ParseThunkSym(record.Data)
		if err != nil {
			return nil
		}
		return &ThunkSymbol{
			baseSymbol: baseSymbol{name: thunk.Name},
			section:    thunk.Segment,
			offset:     thunk.Offset,
			length:     uint32(thunk.Length),
		}

	default:
		return nil
	}
}

// SymbolCount returns the number of symbols in this module.
func (m *Module) SymbolCount() int {
	m.loadSymbols()
	if m.symbolsErr != nil {
		return 0
	}
	return len(m.symbols)
}

// LocalSymbol represents a local variable.
type LocalSymbol struct {
	baseSymbol
	typeIndex   uint32
	isParameter bool
}

func (s *LocalSymbol) Kind() SymbolKind {
	if s.isParameter {
		return SymbolKindParameter
	}
	return SymbolKindLocal
}

func (s *LocalSymbol) Section() uint16   { return 0 }
func (s *LocalSymbol) Offset() uint32    { return 0 }
func (s *LocalSymbol) TypeIndex() uint32 { return s.typeIndex }
func (s *LocalSymbol) IsParameter() bool { return s.isParameter }

// LabelSymbol represents a code label.
type LabelSymbol struct {
	baseSymbol
	section uint16
	offset  uint32
}

func (s *LabelSymbol) Kind() SymbolKind { return SymbolKindLabel }
func (s *LabelSymbol) Section() uint16  { return s.section }
func (s *LabelSymbol) Offset() uint32   { return s.offset }

// BlockSymbol represents a code block.
type BlockSymbol struct {
	baseSymbol
	section uint16
	offset  uint32
	length  uint32
}

func (s *BlockSymbol) Kind() SymbolKind { return SymbolKindBlock }
func (s *BlockSymbol) Section() uint16  { return s.section }
func (s *BlockSymbol) Offset() uint32   { return s.offset }
func (s *BlockSymbol) Length() uint32   { return s.length }

// ThunkSymbol represents a thunk.
type ThunkSymbol struct {
	baseSymbol
	section uint16
	offset  uint32
	length  uint32
}

func (s *ThunkSymbol) Kind() SymbolKind { return SymbolKindThunk }
func (s *ThunkSymbol) Section() uint16  { return s.section }
func (s *ThunkSymbol) Offset() uint32   { return s.offset }
func (s *ThunkSymbol) Length() uint32   { return s.length }
