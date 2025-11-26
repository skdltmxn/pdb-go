package tpi

import (
	"errors"
	"fmt"
	"sync"

	"github.com/skdltmxn/pdb-go/internal/stream"
)

// TPI stream version constants
const (
	TPIVersionV40 uint32 = 19950410
	TPIVersionV41 uint32 = 19951122
	TPIVersionV50 uint32 = 19961031
	TPIVersionV70 uint32 = 19990903
	TPIVersionV80 uint32 = 20040203 // Current version
)

// TPI Header size
const TPIHeaderSize = 56

// Errors
var (
	ErrInvalidTPIHeader    = errors.New("tpi: invalid TPI header")
	ErrUnsupportedVersion  = errors.New("tpi: unsupported TPI version")
	ErrTypeIndexOutOfRange = errors.New("tpi: type index out of range")
	ErrInvalidTypeRecord   = errors.New("tpi: invalid type record")
)

// Header represents the TPI or IPI stream header.
type Header struct {
	// Version is always V80 (20040203) in modern PDBs
	Version uint32

	// HeaderSize is the size of this header (typically 56 bytes)
	HeaderSize uint32

	// TypeIndexBegin is the first valid type index (typically 0x1000)
	TypeIndexBegin TypeIndex

	// TypeIndexEnd is one past the last type index
	TypeIndexEnd TypeIndex

	// TypeRecordBytes is the total size of type record data
	TypeRecordBytes uint32

	// HashStreamIndex is the MSF stream containing hash data
	HashStreamIndex uint16

	// HashAuxStreamIndex is auxiliary hash stream (usually 0xFFFF)
	HashAuxStreamIndex uint16

	// HashKeySize is the size of hash keys (typically 4)
	HashKeySize uint32

	// NumHashBuckets is the number of hash buckets
	NumHashBuckets uint32

	// HashValueBufferOffset and HashValueBufferLength describe hash values
	HashValueBufferOffset int32
	HashValueBufferLength uint32

	// IndexOffsetBufferOffset and IndexOffsetBufferLength for type lookups
	IndexOffsetBufferOffset int32
	IndexOffsetBufferLength uint32

	// HashAdjBufferOffset and HashAdjBufferLength for incremental linking
	HashAdjBufferOffset int32
	HashAdjBufferLength uint32
}

// TypeCount returns the number of type records.
func (h *Header) TypeCount() uint32 {
	return uint32(h.TypeIndexEnd - h.TypeIndexBegin)
}

// Stream represents a parsed TPI or IPI stream.
type Stream struct {
	Header Header

	// rawRecords holds the raw type record data
	rawRecords []byte

	// recordOffsets maps TypeIndex to byte offset in rawRecords
	// This enables O(1) random access to types
	recordOffsets map[TypeIndex]uint32

	// parsed types cache with thread-safe access
	typeCache sync.Map // map[TypeIndex]*TypeRecord

	mu sync.RWMutex
}

// ParseStream parses a TPI or IPI stream from raw data.
func ParseStream(data []byte) (*Stream, error) {
	if len(data) < TPIHeaderSize {
		return nil, ErrInvalidTPIHeader
	}

	r := stream.NewReader(data)
	s := &Stream{
		recordOffsets: make(map[TypeIndex]uint32),
	}

	// Parse header
	if err := s.parseHeader(r); err != nil {
		return nil, err
	}

	// Extract raw record data
	recordStart := int(s.Header.HeaderSize)
	recordEnd := recordStart + int(s.Header.TypeRecordBytes)
	if recordEnd > len(data) {
		return nil, fmt.Errorf("tpi: truncated stream: expected %d bytes, got %d", recordEnd, len(data))
	}
	s.rawRecords = data[recordStart:recordEnd]

	// Build offset index for random access
	if err := s.buildOffsetIndex(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Stream) parseHeader(r *stream.Reader) error {
	var err error

	s.Header.Version, err = r.ReadU32()
	if err != nil {
		return err
	}

	// Validate version
	if s.Header.Version != TPIVersionV80 && s.Header.Version != TPIVersionV70 {
		return fmt.Errorf("%w: %d", ErrUnsupportedVersion, s.Header.Version)
	}

	s.Header.HeaderSize, err = r.ReadU32()
	if err != nil {
		return err
	}

	begin, err := r.ReadU32()
	if err != nil {
		return err
	}
	s.Header.TypeIndexBegin = TypeIndex(begin)

	end, err := r.ReadU32()
	if err != nil {
		return err
	}
	s.Header.TypeIndexEnd = TypeIndex(end)

	s.Header.TypeRecordBytes, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.HashStreamIndex, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.HashAuxStreamIndex, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.HashKeySize, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.NumHashBuckets, err = r.ReadU32()
	if err != nil {
		return err
	}

	hashValueOffset, err := r.ReadI32()
	if err != nil {
		return err
	}
	s.Header.HashValueBufferOffset = hashValueOffset

	s.Header.HashValueBufferLength, err = r.ReadU32()
	if err != nil {
		return err
	}

	indexOffsetOffset, err := r.ReadI32()
	if err != nil {
		return err
	}
	s.Header.IndexOffsetBufferOffset = indexOffsetOffset

	s.Header.IndexOffsetBufferLength, err = r.ReadU32()
	if err != nil {
		return err
	}

	hashAdjOffset, err := r.ReadI32()
	if err != nil {
		return err
	}
	s.Header.HashAdjBufferOffset = hashAdjOffset

	s.Header.HashAdjBufferLength, err = r.ReadU32()
	if err != nil {
		return err
	}

	return nil
}

// buildOffsetIndex scans the record data to build the type index -> offset mapping.
func (s *Stream) buildOffsetIndex() error {
	r := stream.NewReader(s.rawRecords)
	typeIndex := s.Header.TypeIndexBegin

	for r.Remaining() > 0 && typeIndex < s.Header.TypeIndexEnd {
		offset := uint32(r.Offset())
		s.recordOffsets[typeIndex] = offset

		// Read record length
		recordLen, err := r.ReadU16()
		if err != nil {
			return err
		}

		// Skip the record data
		if err := r.Skip(int(recordLen)); err != nil {
			return err
		}

		typeIndex++
	}

	return nil
}

// TypeRecord represents a parsed type record.
type TypeRecord struct {
	Kind TypeRecordKind
	Data []byte // Raw record data (excluding length and kind)
}

// GetTypeRecord returns the raw type record for the given index.
func (s *Stream) GetTypeRecord(ti TypeIndex) (*TypeRecord, error) {
	// Check cache first
	if cached, ok := s.typeCache.Load(ti); ok {
		return cached.(*TypeRecord), nil
	}

	// Check for simple type
	if ti.IsSimpleType() {
		return nil, nil // Simple types don't have records
	}

	// Check range
	if ti < s.Header.TypeIndexBegin || ti >= s.Header.TypeIndexEnd {
		return nil, fmt.Errorf("%w: %d", ErrTypeIndexOutOfRange, ti)
	}

	// Get offset
	offset, ok := s.recordOffsets[ti]
	if !ok {
		return nil, fmt.Errorf("%w: no offset for type %d", ErrTypeIndexOutOfRange, ti)
	}

	// Parse record
	r := stream.NewReader(s.rawRecords[offset:])

	recordLen, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	kind, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	// recordLen includes the kind field, so subtract 2
	dataLen := int(recordLen) - 2
	if dataLen < 0 {
		return nil, ErrInvalidTypeRecord
	}

	data, err := r.ReadBytesRef(dataLen)
	if err != nil {
		return nil, err
	}

	record := &TypeRecord{
		Kind: TypeRecordKind(kind),
		Data: data,
	}

	// Cache the result
	s.typeCache.Store(ti, record)

	return record, nil
}

// TypeIndexBegin returns the first valid type index.
func (s *Stream) TypeIndexBegin() TypeIndex {
	return s.Header.TypeIndexBegin
}

// TypeIndexEnd returns one past the last valid type index.
func (s *Stream) TypeIndexEnd() TypeIndex {
	return s.Header.TypeIndexEnd
}

// TypeCount returns the number of type records.
func (s *Stream) TypeCount() uint32 {
	return s.Header.TypeCount()
}

// ModifierRecord represents an LF_MODIFIER type.
type ModifierRecord struct {
	ModifiedType TypeIndex
	Modifiers    ModifierOptions
}

// ParseModifierRecord parses an LF_MODIFIER record.
func ParseModifierRecord(data []byte) (*ModifierRecord, error) {
	r := stream.NewReader(data)

	modType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	mods, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	return &ModifierRecord{
		ModifiedType: TypeIndex(modType),
		Modifiers:    ModifierOptions(mods),
	}, nil
}

// PointerRecord represents an LF_POINTER type.
type PointerRecord struct {
	ReferentType TypeIndex
	Attributes   PointerAttributes
	// MemberInfo is present only for pointer-to-member
	ContainingClass TypeIndex // Only if pointer-to-member
}

// ParsePointerRecord parses an LF_POINTER record.
func ParsePointerRecord(data []byte) (*PointerRecord, error) {
	r := stream.NewReader(data)

	refType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	attrs, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	rec := &PointerRecord{
		ReferentType: TypeIndex(refType),
		Attributes:   PointerAttributes(attrs),
	}

	// Check if this is a pointer-to-member
	mode := rec.Attributes.Mode()
	if mode == PointerModePointerToDataMember || mode == PointerModePointerToMemberFunction {
		containingClass, err := r.ReadU32()
		if err != nil {
			return nil, err
		}
		rec.ContainingClass = TypeIndex(containingClass)
	}

	return rec, nil
}

// ProcedureRecord represents an LF_PROCEDURE type (function signature).
type ProcedureRecord struct {
	ReturnType      TypeIndex
	CallingConv     CallingConvention
	FunctionOptions FunctionOptions
	ParameterCount  uint16
	ArgumentList    TypeIndex
}

// ParseProcedureRecord parses an LF_PROCEDURE record.
func ParseProcedureRecord(data []byte) (*ProcedureRecord, error) {
	r := stream.NewReader(data)

	retType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	callConv, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	funcOpts, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	paramCount, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	argList, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	return &ProcedureRecord{
		ReturnType:      TypeIndex(retType),
		CallingConv:     CallingConvention(callConv),
		FunctionOptions: FunctionOptions(funcOpts),
		ParameterCount:  paramCount,
		ArgumentList:    TypeIndex(argList),
	}, nil
}

// MFunctionRecord represents an LF_MFUNCTION type (member function).
type MFunctionRecord struct {
	ReturnType      TypeIndex
	ClassType       TypeIndex
	ThisType        TypeIndex
	CallingConv     CallingConvention
	FunctionOptions FunctionOptions
	ParameterCount  uint16
	ArgumentList    TypeIndex
	ThisAdjust      int32
}

// ParseMFunctionRecord parses an LF_MFUNCTION record.
func ParseMFunctionRecord(data []byte) (*MFunctionRecord, error) {
	r := stream.NewReader(data)

	retType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	classType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	thisType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	callConv, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	funcOpts, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	paramCount, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	argList, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	thisAdjust, err := r.ReadI32()
	if err != nil {
		return nil, err
	}

	return &MFunctionRecord{
		ReturnType:      TypeIndex(retType),
		ClassType:       TypeIndex(classType),
		ThisType:        TypeIndex(thisType),
		CallingConv:     CallingConvention(callConv),
		FunctionOptions: FunctionOptions(funcOpts),
		ParameterCount:  paramCount,
		ArgumentList:    TypeIndex(argList),
		ThisAdjust:      thisAdjust,
	}, nil
}

// ArgListRecord represents an LF_ARGLIST type.
type ArgListRecord struct {
	ArgTypes []TypeIndex
}

// ParseArgListRecord parses an LF_ARGLIST record.
func ParseArgListRecord(data []byte) (*ArgListRecord, error) {
	r := stream.NewReader(data)

	count, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	args := make([]TypeIndex, count)
	for i := uint32(0); i < count; i++ {
		argType, err := r.ReadU32()
		if err != nil {
			return nil, err
		}
		args[i] = TypeIndex(argType)
	}

	return &ArgListRecord{ArgTypes: args}, nil
}

// ArrayRecord represents an LF_ARRAY type.
type ArrayRecord struct {
	ElementType TypeIndex
	IndexType   TypeIndex
	Size        uint64
	Name        string
}

// ParseArrayRecord parses an LF_ARRAY record.
func ParseArrayRecord(data []byte) (*ArrayRecord, error) {
	r := stream.NewReader(data)

	elemType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	indexType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	size, err := r.ReadNumeric()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	return &ArrayRecord{
		ElementType: TypeIndex(elemType),
		IndexType:   TypeIndex(indexType),
		Size:        size,
		Name:        name,
	}, nil
}

// ClassRecord represents an LF_CLASS, LF_STRUCTURE, or LF_INTERFACE type.
type ClassRecord struct {
	MemberCount uint16
	Properties  ClassProperties
	FieldList   TypeIndex
	DerivedFrom TypeIndex
	VShape      TypeIndex
	Size        uint64
	Name        string
	UniqueName  string // Only if HasUniqueName property is set
}

// ParseClassRecord parses an LF_CLASS, LF_STRUCTURE, or LF_INTERFACE record.
func ParseClassRecord(data []byte) (*ClassRecord, error) {
	r := stream.NewReader(data)

	memberCount, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	props, err := r.ReadU16()
	if err != nil {
		return nil, err
	}
	properties := ClassProperties(props)

	fieldList, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	derivedFrom, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	vshape, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	size, err := r.ReadNumeric()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	rec := &ClassRecord{
		MemberCount: memberCount,
		Properties:  properties,
		FieldList:   TypeIndex(fieldList),
		DerivedFrom: TypeIndex(derivedFrom),
		VShape:      TypeIndex(vshape),
		Size:        size,
		Name:        name,
	}

	// Read unique name if present
	if properties.HasUniqueName() {
		uniqueName, err := r.ReadCString()
		if err != nil {
			return nil, err
		}
		rec.UniqueName = uniqueName
	}

	return rec, nil
}

// UnionRecord represents an LF_UNION type.
type UnionRecord struct {
	MemberCount uint16
	Properties  ClassProperties
	FieldList   TypeIndex
	Size        uint64
	Name        string
	UniqueName  string
}

// ParseUnionRecord parses an LF_UNION record.
func ParseUnionRecord(data []byte) (*UnionRecord, error) {
	r := stream.NewReader(data)

	memberCount, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	props, err := r.ReadU16()
	if err != nil {
		return nil, err
	}
	properties := ClassProperties(props)

	fieldList, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	size, err := r.ReadNumeric()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	rec := &UnionRecord{
		MemberCount: memberCount,
		Properties:  properties,
		FieldList:   TypeIndex(fieldList),
		Size:        size,
		Name:        name,
	}

	if properties.HasUniqueName() {
		uniqueName, err := r.ReadCString()
		if err != nil {
			return nil, err
		}
		rec.UniqueName = uniqueName
	}

	return rec, nil
}

// EnumRecord represents an LF_ENUM type.
type EnumRecord struct {
	Count          uint16
	Properties     ClassProperties
	UnderlyingType TypeIndex
	FieldList      TypeIndex
	Name           string
	UniqueName     string
}

// ParseEnumRecord parses an LF_ENUM record.
func ParseEnumRecord(data []byte) (*EnumRecord, error) {
	r := stream.NewReader(data)

	count, err := r.ReadU16()
	if err != nil {
		return nil, err
	}

	props, err := r.ReadU16()
	if err != nil {
		return nil, err
	}
	properties := ClassProperties(props)

	underlyingType, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	fieldList, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	name, err := r.ReadCString()
	if err != nil {
		return nil, err
	}

	rec := &EnumRecord{
		Count:          count,
		Properties:     properties,
		UnderlyingType: TypeIndex(underlyingType),
		FieldList:      TypeIndex(fieldList),
		Name:           name,
	}

	if properties.HasUniqueName() {
		uniqueName, err := r.ReadCString()
		if err != nil {
			return nil, err
		}
		rec.UniqueName = uniqueName
	}

	return rec, nil
}

// BitFieldRecord represents an LF_BITFIELD type.
type BitFieldRecord struct {
	Type     TypeIndex
	Length   uint8
	Position uint8
}

// ParseBitFieldRecord parses an LF_BITFIELD record.
func ParseBitFieldRecord(data []byte) (*BitFieldRecord, error) {
	r := stream.NewReader(data)

	typ, err := r.ReadU32()
	if err != nil {
		return nil, err
	}

	length, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	position, err := r.ReadU8()
	if err != nil {
		return nil, err
	}

	return &BitFieldRecord{
		Type:     TypeIndex(typ),
		Length:   length,
		Position: position,
	}, nil
}
