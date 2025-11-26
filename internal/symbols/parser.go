package symbols

import (
	"errors"
	"fmt"

	"github.com/skdltmxn/pdb-go/internal/stream"
	"github.com/skdltmxn/pdb-go/internal/tpi"
)

// Errors
var (
	ErrInvalidSymbolRecord = errors.New("symbols: invalid symbol record")
	ErrUnexpectedEnd       = errors.New("symbols: unexpected end of data")
)

// ParseSymbolRecord parses a single symbol record from raw data.
// Returns the symbol and the number of bytes consumed.
func ParseSymbolRecord(data []byte) (*SymbolRecord, int, error) {
	if len(data) < 4 {
		return nil, 0, ErrUnexpectedEnd
	}

	r := stream.NewReader(data)

	// Read record length (does not include the length field itself)
	length, err := r.ReadU16()
	if err != nil {
		return nil, 0, err
	}

	// Read record kind
	kind, err := r.ReadU16()
	if err != nil {
		return nil, 0, err
	}

	// Total size = 2 (length field) + length
	totalSize := int(length) + 2
	if totalSize > len(data) {
		return nil, 0, ErrUnexpectedEnd
	}

	// Data is everything after the kind field
	dataLen := int(length) - 2
	if dataLen < 0 {
		return nil, 0, ErrInvalidSymbolRecord
	}

	recordData := data[4 : 4+dataLen]

	return &SymbolRecord{
		Kind: SymbolRecordKind(kind),
		Data: recordData,
	}, totalSize, nil
}

// SymbolIterator iterates over symbol records in a stream.
type SymbolIterator struct {
	data   []byte
	offset int
}

// NewSymbolIterator creates a new symbol iterator.
func NewSymbolIterator(data []byte) *SymbolIterator {
	return &SymbolIterator{data: data}
}

// Next returns the next symbol record, or nil if there are no more.
func (it *SymbolIterator) Next() (*SymbolRecord, error) {
	if it.offset >= len(it.data) {
		return nil, nil
	}

	rec, size, err := ParseSymbolRecord(it.data[it.offset:])
	if err != nil {
		return nil, err
	}

	it.offset += size
	return rec, nil
}

// Reset resets the iterator to the beginning.
func (it *SymbolIterator) Reset() {
	it.offset = 0
}

// ParseProcSym parses a procedure symbol (S_GPROC32, S_LPROC32, etc.).
func ParseProcSym(data []byte) (*ProcSym, error) {
	r := stream.NewReader(data)

	ptrParent, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	ptrEnd, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	ptrNext, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	codeSize, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	dbgStart, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	dbgEnd, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	funcType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	codeOffset, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	segment, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	flags, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &ProcSym{
		PtrParent:    ptrParent,
		PtrEnd:       ptrEnd,
		PtrNext:      ptrNext,
		CodeSize:     codeSize,
		DbgStart:     dbgStart,
		DbgEnd:       dbgEnd,
		FunctionType: tpi.TypeIndex(funcType),
		CodeOffset:   codeOffset,
		Segment:      segment,
		Flags:        ProcFlags(flags),
		Name:         name,
	}, nil
}

// ParseDataSym parses a data symbol (S_GDATA32, S_LDATA32, etc.).
func ParseDataSym(data []byte) (*DataSym, error) {
	r := stream.NewReader(data)

	typeIndex, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	offset, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	segment, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &DataSym{
		Type:    tpi.TypeIndex(typeIndex),
		Offset:  offset,
		Segment: segment,
		Name:    name,
	}, nil
}

// ParsePublicSym32 parses a public symbol (S_PUB32).
func ParsePublicSym32(data []byte) (*PublicSym32, error) {
	r := stream.NewReader(data)

	flags, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	offset, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	segment, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &PublicSym32{
		Flags:   PublicSymFlags(flags),
		Offset:  offset,
		Segment: segment,
		Name:    name,
	}, nil
}

// ParseLocalSym parses a local variable symbol (S_LOCAL).
func ParseLocalSym(data []byte) (*LocalSym, error) {
	r := stream.NewReader(data)

	typeIndex, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	flags, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &LocalSym{
		Type:  tpi.TypeIndex(typeIndex),
		Flags: LocalFlags(flags),
		Name:  name,
	}, nil
}

// ParseUDTSym parses a UDT symbol (S_UDT).
func ParseUDTSym(data []byte) (*UDTSym, error) {
	r := stream.NewReader(data)

	typeIndex, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &UDTSym{
		Type: tpi.TypeIndex(typeIndex),
		Name: name,
	}, nil
}

// ParseConstantSym parses a constant symbol (S_CONSTANT).
func ParseConstantSym(data []byte) (*ConstantSym, error) {
	r := stream.NewReader(data)

	typeIndex, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	value, err := r.ReadNumeric()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &ConstantSym{
		Type:  tpi.TypeIndex(typeIndex),
		Value: value,
		Name:  name,
	}, nil
}

// ParseLabelSym parses a label symbol (S_LABEL32).
func ParseLabelSym(data []byte) (*LabelSym, error) {
	r := stream.NewReader(data)

	offset, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	segment, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	flags, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &LabelSym{
		Offset:  offset,
		Segment: segment,
		Flags:   flags,
		Name:    name,
	}, nil
}

// ParseBlockSym parses a block symbol (S_BLOCK32).
func ParseBlockSym(data []byte) (*BlockSym, error) {
	r := stream.NewReader(data)

	ptrParent, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	ptrEnd, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	codeSize, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	offset, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	segment, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &BlockSym{
		PtrParent: ptrParent,
		PtrEnd:    ptrEnd,
		CodeSize:  codeSize,
		Offset:    offset,
		Segment:   segment,
		Name:      name,
	}, nil
}

// ParseThunkSym parses a thunk symbol (S_THUNK32).
func ParseThunkSym(data []byte) (*ThunkSym, error) {
	r := stream.NewReader(data)

	ptrParent, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	ptrEnd, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	ptrNext, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	offset, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	segment, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	length, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	ordinal, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &ThunkSym{
		PtrParent: ptrParent,
		PtrEnd:    ptrEnd,
		PtrNext:   ptrNext,
		Offset:    offset,
		Segment:   segment,
		Length:    length,
		Ordinal:   ordinal,
		Name:      name,
	}, nil
}

// ParseObjNameSym parses an object name symbol (S_OBJNAME).
func ParseObjNameSym(data []byte) (*ObjNameSym, error) {
	r := stream.NewReader(data)

	signature, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &ObjNameSym{
		Signature: signature,
		Name:      name,
	}, nil
}

// ParseCompileSym3 parses a compile symbol (S_COMPILE3).
func ParseCompileSym3(data []byte) (*CompileSym3, error) {
	r := stream.NewReader(data)

	flags, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	machine, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	frontendMajor, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	frontendMinor, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	frontendBuild, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	frontendQFE, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	backendMajor, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	backendMinor, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	backendBuild, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	backendQFE, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	version, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &CompileSym3{
		Flags:         flags,
		Machine:       machine,
		FrontendMajor: frontendMajor,
		FrontendMinor: frontendMinor,
		FrontendBuild: frontendBuild,
		FrontendQFE:   frontendQFE,
		BackendMajor:  backendMajor,
		BackendMinor:  backendMinor,
		BackendBuild:  backendBuild,
		BackendQFE:    backendQFE,
		Version:       version,
	}, nil
}

// ParseRegRelSym parses a register-relative symbol (S_REGREL32).
func ParseRegRelSym(data []byte) (*RegRelSym, error) {
	r := stream.NewReader(data)

	offset, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	typeIndex, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	register, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &RegRelSym{
		Offset:   offset,
		Type:     tpi.TypeIndex(typeIndex),
		Register: register,
		Name:     name,
	}, nil
}

// ParseBPRelSym parses a base-pointer relative symbol (S_BPREL32).
func ParseBPRelSym(data []byte) (*BPRelSym, error) {
	r := stream.NewReader(data)

	offset, err := r.ReadI32()
	if err != nil {
		return nil, err
	}

	typeIndex, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &BPRelSym{
		Offset: offset,
		Type:   tpi.TypeIndex(typeIndex),
		Name:   name,
	}, nil
}

// ParseFrameProcSym parses a frame procedure symbol (S_FRAMEPROC).
func ParseFrameProcSym(data []byte) (*FrameProcSym, error) {
	r := stream.NewReader(data)

	totalFrameBytes, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	paddingFrameBytes, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	offsetToPadding, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	calleeSaveBytes, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	offsetOfExceptionHandler, err := r.ReadI32()
	if err != nil {
		return nil, err
	}

	sectionIdOfExceptionHandler, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	// Skip padding
	_, err = r.ReadU16()
	if err != nil {
		return nil, err
	}

	flags, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	return &FrameProcSym{
		TotalFrameBytes:             totalFrameBytes,
		PaddingFrameBytes:           paddingFrameBytes,
		OffsetToPadding:             offsetToPadding,
		CalleeSaveBytes:             calleeSaveBytes,
		OffsetOfExceptionHandler:    offsetOfExceptionHandler,
		SectionIdOfExceptionHandler: sectionIdOfExceptionHandler,
		Flags:                       flags,
	}, nil
}

// ParseSectionSym parses a section symbol (S_SECTION).
func ParseSectionSym(data []byte) (*SectionSym, error) {
	r := stream.NewReader(data)

	sectionNumber, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	alignment, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	reserved, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	rva, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	length, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	characteristics, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &SectionSym{
		SectionNumber:   sectionNumber,
		Alignment:       alignment,
		Reserved:        reserved,
		RVA:             rva,
		Length:          length,
		Characteristics: characteristics,
		Name:            name,
	}, nil
}

// ParseRefSym parses a reference symbol (S_PROCREF, S_LPROCREF, S_DATAREF).
func ParseRefSym(data []byte) (*RefSym, error) {
	r := stream.NewReader(data)

	sumName, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	ibSym, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	imod, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &RefSym{
		SumName: sumName,
		IBSym:   ibSym,
		Imod:    imod,
		Name:    name,
	}, nil
}

// ParseSymbol parses a symbol record and returns the appropriate typed symbol.
func ParseSymbol(rec *SymbolRecord) (interface{}, error) {
	switch rec.Kind {
	case S_GPROC32, S_LPROC32, S_GPROC32_ID, S_LPROC32_ID:
		return ParseProcSym(rec.Data)
	case S_GDATA32, S_LDATA32, S_GTHREAD32, S_LTHREAD32:
		return ParseDataSym(rec.Data)
	case S_PUB32:
		return ParsePublicSym32(rec.Data)
	case S_LOCAL:
		return ParseLocalSym(rec.Data)
	case S_UDT:
		return ParseUDTSym(rec.Data)
	case S_CONSTANT:
		return ParseConstantSym(rec.Data)
	case S_LABEL32:
		return ParseLabelSym(rec.Data)
	case S_BLOCK32:
		return ParseBlockSym(rec.Data)
	case S_OBJNAME:
		return ParseObjNameSym(rec.Data)
	case S_COMPILE3:
		return ParseCompileSym3(rec.Data)
	case S_REGREL32:
		return ParseRegRelSym(rec.Data)
	case S_BPREL32:
		return ParseBPRelSym(rec.Data)
	case S_FRAMEPROC:
		return ParseFrameProcSym(rec.Data)
	case S_SECTION:
		return ParseSectionSym(rec.Data)
	case S_PROCREF, S_LPROCREF, S_DATAREF:
		return ParseRefSym(rec.Data)
	default:
		// Return the generic record for unsupported types
		return rec, nil
	}
}

// PSIHeader is the header of the Public Symbol Index stream.
type PSIHeader struct {
	// SymHash is the hash of the symbol stream
	SymHash uint32
	// AddrMap size in bytes
	AddrMapSize uint32
	// NumThunks is the number of thunk entries
	NumThunks uint32
	// SizeOfThunk is the size of each thunk entry
	SizeOfThunk uint32
	// ISectThunkTable is the section containing thunks
	ISectThunkTable uint16
	// Padding
	Padding uint16
	// OffThunkTable is the offset of the thunk table
	OffThunkTable uint32
	// NumSects is the number of sections
	NumSects uint32
}

// GSIHeader is the header of the Global Symbol Index stream.
type GSIHeader struct {
	// VersionSignature should be 0xFFFFFFFF
	VersionSignature uint32
	// Version should be 0xeffe0000 + 19990810
	Version uint32
	// HashRecordsSize is the size of hash records in bytes
	HashRecordsSize uint32
	// BucketSize is the size of the bucket array
	BucketSize uint32
}

// HashRecord is an entry in the GSI hash table.
type HashRecord struct {
	// Offset is the offset into the symbol record stream (+1, 0 means empty)
	Offset uint32
	// CRef is the reference count
	CRef uint32
}

// ParsePublicSymbolIndex parses the PSI stream header.
// Returns the address map offsets which point into the symbol record stream.
func ParsePublicSymbolIndex(data []byte) (*PSIHeader, []uint32, error) {
	if len(data) < 28 {
		return nil, nil, fmt.Errorf("symbols: PSI stream too short")
	}

	r := stream.NewReader(data)

	// First comes the GSI header
	var gsiHeader GSIHeader
	var err error

	gsiHeader.VersionSignature, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	gsiHeader.Version, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	gsiHeader.HashRecordsSize, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	gsiHeader.BucketSize, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	// Skip hash records and buckets
	if err := r.Skip(int(gsiHeader.HashRecordsSize)); err != nil {
		return nil, nil, err
	}
	if err := r.Skip(int(gsiHeader.BucketSize)); err != nil {
		return nil, nil, err
	}

	// Now read PSI-specific header
	var psiHeader PSIHeader

	psiHeader.SymHash, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	psiHeader.AddrMapSize, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	psiHeader.NumThunks, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	psiHeader.SizeOfThunk, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	psiHeader.ISectThunkTable, err = r.ReadU16()
	if err != nil {
		return nil, nil, err
	}

	psiHeader.Padding, err = r.ReadU16()
	if err != nil {
		return nil, nil, err
	}

	psiHeader.OffThunkTable, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	psiHeader.NumSects, err = r.ReadU32()
	if err != nil {
		return nil, nil, err
	}

	// Read address map (array of uint32 offsets into symbol record stream)
	numAddrs := psiHeader.AddrMapSize / 4
	addrMap := make([]uint32, numAddrs)
	for i := uint32(0); i < numAddrs; i++ {
		addrMap[i], err = r.ReadU32()
		if err != nil {
			break
		}
	}

	return &psiHeader, addrMap, nil
}

// ParseSymbolRecordStream parses the symbol record stream and extracts public symbols.
// This is the GSS (Global Symbol Stream) referenced by SymRecordStreamIndex.
func ParseSymbolRecordStream(data []byte) ([]*PublicSym32, error) {
	var symbols []*PublicSym32
	r := stream.NewReader(data)

	for r.Remaining() > 4 {
		rec, size, err := ParseSymbolRecord(data[r.Offset():])
		if err != nil {
			break
		}

		if rec.Kind == S_PUB32 {
			sym, err := ParsePublicSym32(rec.Data)
			if err == nil {
				symbols = append(symbols, sym)
			}
		}

		if err := r.Skip(size); err != nil {
			break
		}
	}

	return symbols, nil
}

