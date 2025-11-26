package pdb

import (
	"iter"
	"sync"

	"github.com/skdltmxn/pdb-go/internal/dbi"
	"github.com/skdltmxn/pdb-go/internal/demangle"
	"github.com/skdltmxn/pdb-go/internal/symbols"
)

// SymbolKind identifies the type of symbol.
type SymbolKind uint16

const (
	SymbolKindUnknown SymbolKind = iota
	SymbolKindPublic
	SymbolKindFunction
	SymbolKindData
	SymbolKindLocal
	SymbolKindParameter
	SymbolKindUDT
	SymbolKindConstant
	SymbolKindLabel
	SymbolKindBlock
	SymbolKindThunk
)

func (k SymbolKind) String() string {
	switch k {
	case SymbolKindPublic:
		return "public"
	case SymbolKindFunction:
		return "function"
	case SymbolKindData:
		return "data"
	case SymbolKindLocal:
		return "local"
	case SymbolKindParameter:
		return "parameter"
	case SymbolKindUDT:
		return "udt"
	case SymbolKindConstant:
		return "constant"
	case SymbolKindLabel:
		return "label"
	case SymbolKindBlock:
		return "block"
	case SymbolKindThunk:
		return "thunk"
	default:
		return "unknown"
	}
}

// Symbol is the interface implemented by all symbol types.
type Symbol interface {
	// Name returns the raw (possibly mangled) symbol name.
	Name() string

	// DemangledName returns the demangled name, or the raw name if not mangled.
	DemangledName() string

	// Kind returns the symbol kind.
	Kind() SymbolKind

	// Section returns the section number (1-based, 0 = no section).
	Section() uint16

	// Offset returns the offset within the section.
	Offset() uint32
}

// baseSymbol provides common symbol functionality including lazy demangling.
type baseSymbol struct {
	name          string
	demangledName string
	demangledOnce sync.Once
}

func (s *baseSymbol) Name() string { return s.name }

func (s *baseSymbol) DemangledName() string {
	s.demangledOnce.Do(func() {
		s.demangledName = demangle.DemangleSimple(s.name)
	})
	return s.demangledName
}

// PublicSymbol represents a public symbol export.
type PublicSymbol struct {
	baseSymbol
	section uint16
	offset  uint32
	flags   symbols.PublicSymFlags
}

func (s *PublicSymbol) Kind() SymbolKind { return SymbolKindPublic }
func (s *PublicSymbol) Section() uint16  { return s.section }
func (s *PublicSymbol) Offset() uint32   { return s.offset }
func (s *PublicSymbol) IsCode() bool     { return s.flags.IsCode() }
func (s *PublicSymbol) IsFunction() bool { return s.flags.IsFunction() }

// FunctionSymbol represents a function with full debug info.
type FunctionSymbol struct {
	baseSymbol
	section   uint16
	offset    uint32
	length    uint32
	typeIndex uint32
}

func (s *FunctionSymbol) Kind() SymbolKind  { return SymbolKindFunction }
func (s *FunctionSymbol) Section() uint16   { return s.section }
func (s *FunctionSymbol) Offset() uint32    { return s.offset }
func (s *FunctionSymbol) Length() uint32    { return s.length }
func (s *FunctionSymbol) TypeIndex() uint32 { return s.typeIndex }

// DataSymbol represents a global or static data symbol.
type DataSymbol struct {
	baseSymbol
	section   uint16
	offset    uint32
	typeIndex uint32
}

func (s *DataSymbol) Kind() SymbolKind  { return SymbolKindData }
func (s *DataSymbol) Section() uint16   { return s.section }
func (s *DataSymbol) Offset() uint32    { return s.offset }
func (s *DataSymbol) TypeIndex() uint32 { return s.typeIndex }

// UDTSymbol represents a user-defined type reference.
type UDTSymbol struct {
	baseSymbol
	typeIndex uint32
}

func (s *UDTSymbol) Kind() SymbolKind  { return SymbolKindUDT }
func (s *UDTSymbol) Section() uint16   { return 0 }
func (s *UDTSymbol) Offset() uint32    { return 0 }
func (s *UDTSymbol) TypeIndex() uint32 { return s.typeIndex }

// ConstantSymbol represents a constant.
type ConstantSymbol struct {
	baseSymbol
	value     uint64
	typeIndex uint32
}

func (s *ConstantSymbol) Kind() SymbolKind  { return SymbolKindConstant }
func (s *ConstantSymbol) Section() uint16   { return 0 }
func (s *ConstantSymbol) Offset() uint32    { return 0 }
func (s *ConstantSymbol) Value() uint64     { return s.value }
func (s *ConstantSymbol) TypeIndex() uint32 { return s.typeIndex }

// SymbolTable provides access to symbols in the PDB.
type SymbolTable struct {
	pdb       *File
	dbiStream *dbi.Stream

	// Raw symbol record stream data (lazy-loaded, kept for on-demand parsing)
	symRecordData     []byte
	symRecordDataOnce sync.Once
	symRecordDataErr  error

	// Lazy-loaded public symbols (only populated when iterating all)
	publicSymbols     []*PublicSymbol
	publicSymbolsOnce sync.Once
	publicSymbolsErr  error

	// Fast lookup indices (lazy-built)
	nameIndex     *symbols.NameIndex
	nameIndexOnce sync.Once

	addrIndex     *symbols.AddressIndex
	addrIndexOnce sync.Once

	// PSI for address map
	psi     *symbols.PSI
	psiOnce sync.Once
	psiErr  error

	mu sync.RWMutex
}

func newSymbolTable(pdb *File, dbiStream *dbi.Stream) *SymbolTable {
	return &SymbolTable{
		pdb:       pdb,
		dbiStream: dbiStream,
	}
}

// ensureSymRecordData loads the symbol record stream data.
func (st *SymbolTable) ensureSymRecordData() error {
	st.symRecordDataOnce.Do(func() {
		if st.dbiStream.Header.SymRecordStreamIndex == 0xFFFF {
			return
		}
		st.symRecordData, st.symRecordDataErr = st.pdb.msf.ReadStream(
			uint32(st.dbiStream.Header.SymRecordStreamIndex))
	})
	return st.symRecordDataErr
}

// ensurePSI loads and parses the PSI stream.
func (st *SymbolTable) ensurePSI() error {
	st.psiOnce.Do(func() {
		if st.dbiStream.Header.PublicStreamIndex == 0xFFFF {
			return
		}
		data, err := st.pdb.msf.ReadStream(uint32(st.dbiStream.Header.PublicStreamIndex))
		if err != nil {
			st.psiErr = err
			return
		}
		st.psi, st.psiErr = symbols.ParsePSI(data)
	})
	return st.psiErr
}

// All returns an iterator over all symbols.
func (st *SymbolTable) All() iter.Seq[Symbol] {
	return func(yield func(Symbol) bool) {
		// First yield public symbols
		for sym := range st.Public() {
			if !yield(sym) {
				return
			}
		}

		// Then yield module symbols
		modules, err := st.pdb.Modules()
		if err != nil {
			return
		}

		for _, mod := range modules {
			for sym := range mod.Symbols() {
				if !yield(sym) {
					return
				}
			}
		}
	}
}

// Public returns an iterator over public symbols only.
// This streams symbols on-demand without loading all into memory.
func (st *SymbolTable) Public() iter.Seq[*PublicSymbol] {
	return func(yield func(*PublicSymbol) bool) {
		if err := st.ensureSymRecordData(); err != nil || st.symRecordData == nil {
			return
		}

		// Stream through symbol records without pre-loading all
		data := st.symRecordData
		offset := 0

		for offset < len(data)-4 {
			rec, size, err := symbols.ParseSymbolRecord(data[offset:])
			if err != nil {
				break
			}

			if rec.Kind == symbols.S_PUB32 {
				sym, err := symbols.ParsePublicSym32(rec.Data)
				if err == nil {
					pubSym := &PublicSymbol{
						baseSymbol: baseSymbol{name: sym.Name},
						section:    sym.Segment,
						offset:     sym.Offset,
						flags:      sym.Flags,
					}
					if !yield(pubSym) {
						return
					}
				}
			}

			offset += size
		}
	}
}

// PublicCached returns all public symbols, caching them for repeated access.
// Use this when you need to iterate multiple times over public symbols.
func (st *SymbolTable) PublicCached() ([]*PublicSymbol, error) {
	st.publicSymbolsOnce.Do(func() {
		st.publicSymbols, st.publicSymbolsErr = st.loadPublicSymbols()
	})
	return st.publicSymbols, st.publicSymbolsErr
}

func (st *SymbolTable) loadPublicSymbols() ([]*PublicSymbol, error) {
	if err := st.ensureSymRecordData(); err != nil {
		return nil, err
	}
	if st.symRecordData == nil {
		return nil, nil
	}

	// Count first to pre-allocate (single pass overhead but better allocation)
	data := st.symRecordData
	count := 0
	offset := 0
	for offset < len(data)-4 {
		rec, size, err := symbols.ParseSymbolRecord(data[offset:])
		if err != nil {
			break
		}
		if rec.Kind == symbols.S_PUB32 {
			count++
		}
		offset += size
	}

	// Second pass to collect
	result := make([]*PublicSymbol, 0, count)
	offset = 0
	for offset < len(data)-4 {
		rec, size, err := symbols.ParseSymbolRecord(data[offset:])
		if err != nil {
			break
		}

		if rec.Kind == symbols.S_PUB32 {
			sym, err := symbols.ParsePublicSym32(rec.Data)
			if err == nil {
				result = append(result, &PublicSymbol{
					baseSymbol: baseSymbol{name: sym.Name},
					section:    sym.Segment,
					offset:     sym.Offset,
					flags:      sym.Flags,
				})
			}
		}

		offset += size
	}

	return result, nil
}

// ByName looks up symbols by their (possibly mangled) name.
// Uses hash-based index for O(1) average lookup.
func (st *SymbolTable) ByName(name string) iter.Seq[Symbol] {
	return func(yield func(Symbol) bool) {
		st.buildNameIndex()

		if st.nameIndex == nil {
			return
		}

		// Find symbol offsets by name
		offsets := st.nameIndex.FindByName(name)
		for _, offset := range offsets {
			sym := st.parseSymbolAt(offset)
			if sym != nil {
				if !yield(sym) {
					return
				}
			}
		}
	}
}

// FindByName finds the first symbol with the given name.
// This is faster than ByName when you only need one result.
func (st *SymbolTable) FindByName(name string) (Symbol, bool) {
	st.buildNameIndex()

	if st.nameIndex == nil {
		return nil, false
	}

	offsets := st.nameIndex.FindByName(name)
	if len(offsets) == 0 {
		return nil, false
	}

	sym := st.parseSymbolAt(offsets[0])
	return sym, sym != nil
}

func (st *SymbolTable) buildNameIndex() {
	st.nameIndexOnce.Do(func() {
		if err := st.ensureSymRecordData(); err != nil || st.symRecordData == nil {
			return
		}
		st.nameIndex = symbols.NewNameIndex(st.symRecordData)
	})
}

// ByAddress looks up symbols containing the given address.
// Uses binary search on address-sorted index for O(log n) lookup.
func (st *SymbolTable) ByAddress(section uint16, offset uint32) (Symbol, bool) {
	st.buildAddrIndex()

	if st.addrIndex == nil {
		return nil, false
	}

	symOffset, _, found := st.addrIndex.FindByAddress(section, offset)
	if !found {
		return nil, false
	}

	sym := st.parseSymbolAt(symOffset)
	return sym, sym != nil
}

// FindSymbolContaining finds the symbol that contains the given address.
// This is useful for finding which function an address belongs to.
func (st *SymbolTable) FindSymbolContaining(section uint16, offset uint32) (Symbol, bool) {
	st.buildAddrIndex()

	if st.addrIndex == nil {
		return nil, false
	}

	symOffset, _, found := st.addrIndex.FindByAddress(section, offset)
	if !found {
		return nil, false
	}

	sym := st.parseSymbolAt(symOffset)
	return sym, sym != nil
}

func (st *SymbolTable) buildAddrIndex() {
	st.addrIndexOnce.Do(func() {
		if err := st.ensureSymRecordData(); err != nil || st.symRecordData == nil {
			return
		}
		if err := st.ensurePSI(); err != nil || st.psi == nil {
			return
		}
		st.addrIndex = symbols.NewAddressIndex(st.psi.AddressMap(), st.symRecordData)
	})
}

// parseSymbolAt parses a symbol at the given offset in the symbol record stream.
func (st *SymbolTable) parseSymbolAt(offset uint32) Symbol {
	if st.symRecordData == nil || int(offset) >= len(st.symRecordData) {
		return nil
	}

	rec, _, err := symbols.ParseSymbolRecord(st.symRecordData[offset:])
	if err != nil {
		return nil
	}

	return st.convertSymbolRecord(rec)
}

func (st *SymbolTable) convertSymbolRecord(rec *symbols.SymbolRecord) Symbol {
	switch rec.Kind {
	case symbols.S_PUB32:
		sym, err := symbols.ParsePublicSym32(rec.Data)
		if err != nil {
			return nil
		}
		return &PublicSymbol{
			baseSymbol: baseSymbol{name: sym.Name},
			section:    sym.Segment,
			offset:     sym.Offset,
			flags:      sym.Flags,
		}

	case symbols.S_GPROC32, symbols.S_LPROC32, symbols.S_GPROC32_ID, symbols.S_LPROC32_ID:
		sym, err := symbols.ParseProcSym(rec.Data)
		if err != nil {
			return nil
		}
		return &FunctionSymbol{
			baseSymbol: baseSymbol{name: sym.Name},
			section:    sym.Segment,
			offset:     sym.CodeOffset,
			length:     sym.CodeSize,
			typeIndex:  uint32(sym.FunctionType),
		}

	case symbols.S_GDATA32, symbols.S_LDATA32:
		sym, err := symbols.ParseDataSym(rec.Data)
		if err != nil {
			return nil
		}
		return &DataSymbol{
			baseSymbol: baseSymbol{name: sym.Name},
			section:    sym.Segment,
			offset:     sym.Offset,
			typeIndex:  uint32(sym.Type),
		}

	case symbols.S_UDT:
		sym, err := symbols.ParseUDTSym(rec.Data)
		if err != nil {
			return nil
		}
		return &UDTSymbol{
			baseSymbol: baseSymbol{name: sym.Name},
			typeIndex:  uint32(sym.Type),
		}

	case symbols.S_CONSTANT:
		sym, err := symbols.ParseConstantSym(rec.Data)
		if err != nil {
			return nil
		}
		return &ConstantSymbol{
			baseSymbol: baseSymbol{name: sym.Name},
			value:      sym.Value,
			typeIndex:  uint32(sym.Type),
		}

	default:
		return nil
	}
}

// Count returns the total number of symbols.
func (st *SymbolTable) Count() int {
	count := 0
	for range st.All() {
		count++
	}
	return count
}

// PublicCount returns the number of public symbols.
// This counts without fully parsing or caching symbols.
func (st *SymbolTable) PublicCount() int {
	// If already cached, use that
	if st.publicSymbols != nil {
		return len(st.publicSymbols)
	}

	// Otherwise count without caching
	if err := st.ensureSymRecordData(); err != nil || st.symRecordData == nil {
		return 0
	}

	count := 0
	data := st.symRecordData
	offset := 0
	for offset < len(data)-4 {
		rec, size, err := symbols.ParseSymbolRecord(data[offset:])
		if err != nil {
			break
		}
		if rec.Kind == symbols.S_PUB32 {
			count++
		}
		offset += size
	}
	return count
}
