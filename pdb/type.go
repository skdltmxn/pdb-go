package pdb

import (
	"iter"
	"sync"

	"github.com/skdltmxn/pdb-go/internal/tpi"
)

// TypeKind identifies the category of a type.
type TypeKind uint16

const (
	TypeKindUnknown TypeKind = iota
	TypeKindPrimitive
	TypeKindPointer
	TypeKindArray
	TypeKindFunction
	TypeKindMemberFunction
	TypeKindClass
	TypeKindStruct
	TypeKindUnion
	TypeKindEnum
	TypeKindBitfield
	TypeKindModifier
	TypeKindArgList
	TypeKindFieldList
)

func (k TypeKind) String() string {
	switch k {
	case TypeKindPrimitive:
		return "primitive"
	case TypeKindPointer:
		return "pointer"
	case TypeKindArray:
		return "array"
	case TypeKindFunction:
		return "function"
	case TypeKindMemberFunction:
		return "member_function"
	case TypeKindClass:
		return "class"
	case TypeKindStruct:
		return "struct"
	case TypeKindUnion:
		return "union"
	case TypeKindEnum:
		return "enum"
	case TypeKindBitfield:
		return "bitfield"
	case TypeKindModifier:
		return "modifier"
	case TypeKindArgList:
		return "arglist"
	case TypeKindFieldList:
		return "fieldlist"
	default:
		return "unknown"
	}
}

// TypeIndex is a reference to a type in the type table.
type TypeIndex uint32

// IsSimpleType returns true if this is a built-in primitive type.
func (ti TypeIndex) IsSimpleType() bool {
	return tpi.TypeIndex(ti).IsSimpleType()
}

// Type provides information about a type record.
type Type interface {
	// Index returns the type index.
	Index() TypeIndex

	// Kind returns the type kind.
	Kind() TypeKind

	// Name returns the type name (if any).
	Name() string

	// Size returns the size in bytes (0 if unknown).
	Size() uint64
}

// PrimitiveType represents a built-in type.
type PrimitiveType struct {
	index     TypeIndex
	name      string
	size      uint64
	isPointer bool
}

func (t *PrimitiveType) Index() TypeIndex { return t.index }
func (t *PrimitiveType) Kind() TypeKind   { return TypeKindPrimitive }
func (t *PrimitiveType) Name() string     { return t.name }
func (t *PrimitiveType) Size() uint64     { return t.size }
func (t *PrimitiveType) IsPointer() bool  { return t.isPointer }

// PointerType represents a pointer type.
type PointerType struct {
	index        TypeIndex
	referentType TypeIndex
	size         uint64
	isConst      bool
	isVolatile   bool
	isReference  bool
	isRValue     bool
}

func (t *PointerType) Index() TypeIndex    { return t.index }
func (t *PointerType) Kind() TypeKind      { return TypeKindPointer }
func (t *PointerType) Name() string        { return "" }
func (t *PointerType) Size() uint64        { return t.size }
func (t *PointerType) ReferentType() TypeIndex { return t.referentType }
func (t *PointerType) IsConst() bool       { return t.isConst }
func (t *PointerType) IsVolatile() bool    { return t.isVolatile }
func (t *PointerType) IsReference() bool   { return t.isReference }
func (t *PointerType) IsRValueRef() bool   { return t.isRValue }

// ArrayType represents an array type.
type ArrayType struct {
	index       TypeIndex
	elementType TypeIndex
	indexType   TypeIndex
	size        uint64
	name        string
}

func (t *ArrayType) Index() TypeIndex      { return t.index }
func (t *ArrayType) Kind() TypeKind        { return TypeKindArray }
func (t *ArrayType) Name() string          { return t.name }
func (t *ArrayType) Size() uint64          { return t.size }
func (t *ArrayType) ElementType() TypeIndex { return t.elementType }
func (t *ArrayType) IndexType() TypeIndex  { return t.indexType }

// FunctionType represents a function signature.
type FunctionType struct {
	index           TypeIndex
	returnType      TypeIndex
	argumentList    TypeIndex
	callingConv     string
	parameterCount  uint16
	isVariadic      bool
}

func (t *FunctionType) Index() TypeIndex     { return t.index }
func (t *FunctionType) Kind() TypeKind       { return TypeKindFunction }
func (t *FunctionType) Name() string         { return "" }
func (t *FunctionType) Size() uint64         { return 0 }
func (t *FunctionType) ReturnType() TypeIndex { return t.returnType }
func (t *FunctionType) ArgumentList() TypeIndex { return t.argumentList }
func (t *FunctionType) CallingConvention() string { return t.callingConv }
func (t *FunctionType) ParameterCount() uint16 { return t.parameterCount }

// MemberFunctionType represents a member function signature.
type MemberFunctionType struct {
	index          TypeIndex
	returnType     TypeIndex
	classType      TypeIndex
	thisType       TypeIndex
	argumentList   TypeIndex
	callingConv    string
	parameterCount uint16
	thisAdjust     int32
}

func (t *MemberFunctionType) Index() TypeIndex     { return t.index }
func (t *MemberFunctionType) Kind() TypeKind       { return TypeKindMemberFunction }
func (t *MemberFunctionType) Name() string         { return "" }
func (t *MemberFunctionType) Size() uint64         { return 0 }
func (t *MemberFunctionType) ReturnType() TypeIndex { return t.returnType }
func (t *MemberFunctionType) ClassType() TypeIndex { return t.classType }
func (t *MemberFunctionType) ThisType() TypeIndex  { return t.thisType }
func (t *MemberFunctionType) ArgumentList() TypeIndex { return t.argumentList }
func (t *MemberFunctionType) CallingConvention() string { return t.callingConv }
func (t *MemberFunctionType) ParameterCount() uint16 { return t.parameterCount }
func (t *MemberFunctionType) ThisAdjust() int32    { return t.thisAdjust }

// ClassType represents a class type.
type ClassType struct {
	index       TypeIndex
	name        string
	uniqueName  string
	size        uint64
	memberCount uint16
	fieldList   TypeIndex
	derivedFrom TypeIndex
	vshape      TypeIndex
	isForwardRef bool
}

func (t *ClassType) Index() TypeIndex      { return t.index }
func (t *ClassType) Kind() TypeKind        { return TypeKindClass }
func (t *ClassType) Name() string          { return t.name }
func (t *ClassType) Size() uint64          { return t.size }
func (t *ClassType) UniqueName() string    { return t.uniqueName }
func (t *ClassType) MemberCount() uint16   { return t.memberCount }
func (t *ClassType) FieldList() TypeIndex  { return t.fieldList }
func (t *ClassType) DerivedFrom() TypeIndex { return t.derivedFrom }
func (t *ClassType) VShape() TypeIndex     { return t.vshape }
func (t *ClassType) IsForwardRef() bool    { return t.isForwardRef }

// StructType represents a struct type.
type StructType struct {
	index       TypeIndex
	name        string
	uniqueName  string
	size        uint64
	memberCount uint16
	fieldList   TypeIndex
	derivedFrom TypeIndex
	vshape      TypeIndex
	isForwardRef bool
}

func (t *StructType) Index() TypeIndex      { return t.index }
func (t *StructType) Kind() TypeKind        { return TypeKindStruct }
func (t *StructType) Name() string          { return t.name }
func (t *StructType) Size() uint64          { return t.size }
func (t *StructType) UniqueName() string    { return t.uniqueName }
func (t *StructType) MemberCount() uint16   { return t.memberCount }
func (t *StructType) FieldList() TypeIndex  { return t.fieldList }
func (t *StructType) DerivedFrom() TypeIndex { return t.derivedFrom }
func (t *StructType) VShape() TypeIndex     { return t.vshape }
func (t *StructType) IsForwardRef() bool    { return t.isForwardRef }

// UnionType represents a union type.
type UnionType struct {
	index       TypeIndex
	name        string
	uniqueName  string
	size        uint64
	memberCount uint16
	fieldList   TypeIndex
	isForwardRef bool
}

func (t *UnionType) Index() TypeIndex     { return t.index }
func (t *UnionType) Kind() TypeKind       { return TypeKindUnion }
func (t *UnionType) Name() string         { return t.name }
func (t *UnionType) Size() uint64         { return t.size }
func (t *UnionType) UniqueName() string   { return t.uniqueName }
func (t *UnionType) MemberCount() uint16  { return t.memberCount }
func (t *UnionType) FieldList() TypeIndex { return t.fieldList }
func (t *UnionType) IsForwardRef() bool   { return t.isForwardRef }

// EnumType represents an enum type.
type EnumType struct {
	index          TypeIndex
	name           string
	uniqueName     string
	underlyingType TypeIndex
	fieldList      TypeIndex
	count          uint16
	isForwardRef   bool
}

func (t *EnumType) Index() TypeIndex         { return t.index }
func (t *EnumType) Kind() TypeKind           { return TypeKindEnum }
func (t *EnumType) Name() string             { return t.name }
func (t *EnumType) Size() uint64             { return 0 } // Size depends on underlying type
func (t *EnumType) UniqueName() string       { return t.uniqueName }
func (t *EnumType) UnderlyingType() TypeIndex { return t.underlyingType }
func (t *EnumType) FieldList() TypeIndex     { return t.fieldList }
func (t *EnumType) Count() uint16            { return t.count }
func (t *EnumType) IsForwardRef() bool       { return t.isForwardRef }

// BitfieldType represents a bitfield type.
type BitfieldType struct {
	index       TypeIndex
	underlyingType TypeIndex
	length      uint8
	position    uint8
}

func (t *BitfieldType) Index() TypeIndex         { return t.index }
func (t *BitfieldType) Kind() TypeKind           { return TypeKindBitfield }
func (t *BitfieldType) Name() string             { return "" }
func (t *BitfieldType) Size() uint64             { return 0 }
func (t *BitfieldType) UnderlyingType() TypeIndex { return t.underlyingType }
func (t *BitfieldType) Length() uint8            { return t.length }
func (t *BitfieldType) Position() uint8          { return t.position }

// ModifierType represents a modified type (const, volatile, etc.).
type ModifierType struct {
	index        TypeIndex
	modifiedType TypeIndex
	isConst      bool
	isVolatile   bool
	isUnaligned  bool
}

func (t *ModifierType) Index() TypeIndex       { return t.index }
func (t *ModifierType) Kind() TypeKind         { return TypeKindModifier }
func (t *ModifierType) Name() string           { return "" }
func (t *ModifierType) Size() uint64           { return 0 }
func (t *ModifierType) ModifiedType() TypeIndex { return t.modifiedType }
func (t *ModifierType) IsConst() bool          { return t.isConst }
func (t *ModifierType) IsVolatile() bool       { return t.isVolatile }
func (t *ModifierType) IsUnaligned() bool      { return t.isUnaligned }

// TypeTable provides access to types in the PDB.
type TypeTable struct {
	tpiStream *tpi.Stream

	// Lazy-loaded types
	typeCache sync.Map // map[TypeIndex]Type

	// Index by name for named types
	byName     map[string][]Type
	byNameOnce sync.Once

	mu sync.RWMutex
}

func newTypeTable(tpiStream *tpi.Stream) *TypeTable {
	return &TypeTable{
		tpiStream: tpiStream,
	}
}

// All returns an iterator over all types.
func (tt *TypeTable) All() iter.Seq[Type] {
	return func(yield func(Type) bool) {
		begin := tt.tpiStream.TypeIndexBegin()
		end := tt.tpiStream.TypeIndexEnd()

		for ti := begin; ti < end; ti++ {
			typ, err := tt.ByIndex(TypeIndex(ti))
			if err != nil || typ == nil {
				continue
			}
			if !yield(typ) {
				return
			}
		}
	}
}

// ByIndex returns the type at the given index.
func (tt *TypeTable) ByIndex(index TypeIndex) (Type, error) {
	// Check cache
	if cached, ok := tt.typeCache.Load(index); ok {
		return cached.(Type), nil
	}

	// Handle simple types
	if index.IsSimpleType() {
		typ := tt.parseSimpleType(index)
		tt.typeCache.Store(index, typ)
		return typ, nil
	}

	// Get the type record
	record, err := tt.tpiStream.GetTypeRecord(tpi.TypeIndex(index))
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrTypeNotFound
	}

	// Parse the type
	typ, err := tt.parseTypeRecord(index, record)
	if err != nil {
		return nil, err
	}

	// Cache and return
	tt.typeCache.Store(index, typ)
	return typ, nil
}

// ByName looks up types by name.
func (tt *TypeTable) ByName(name string) iter.Seq[Type] {
	return func(yield func(Type) bool) {
		tt.buildNameIndex()

		tt.mu.RLock()
		types := tt.byName[name]
		tt.mu.RUnlock()

		for _, typ := range types {
			if !yield(typ) {
				return
			}
		}
	}
}

func (tt *TypeTable) buildNameIndex() {
	tt.byNameOnce.Do(func() {
		tt.mu.Lock()
		defer tt.mu.Unlock()

		tt.byName = make(map[string][]Type)

		for typ := range tt.All() {
			name := typ.Name()
			if name != "" {
				tt.byName[name] = append(tt.byName[name], typ)
			}
		}
	})
}

// Count returns the total number of types.
func (tt *TypeTable) Count() uint32 {
	return tt.tpiStream.TypeCount()
}

// FirstIndex returns the first valid type index.
func (tt *TypeTable) FirstIndex() TypeIndex {
	return TypeIndex(tt.tpiStream.TypeIndexBegin())
}

// LastIndex returns the last valid type index.
func (tt *TypeTable) LastIndex() TypeIndex {
	return TypeIndex(tt.tpiStream.TypeIndexEnd() - 1)
}

func (tt *TypeTable) parseSimpleType(index TypeIndex) Type {
	ti := tpi.TypeIndex(index)
	kind := ti.SimpleKind()
	mode := ti.SimpleMode()

	var name string
	var size uint64

	switch kind {
	case tpi.SimpleTypeVoid:
		name = "void"
		size = 0
	case tpi.SimpleTypeSignedChar:
		name = "signed char"
		size = 1
	case tpi.SimpleTypeUnsignedChar:
		name = "unsigned char"
		size = 1
	case tpi.SimpleTypeNarrowChar:
		name = "char"
		size = 1
	case tpi.SimpleTypeWideChar:
		name = "wchar_t"
		size = 2
	case tpi.SimpleTypeChar16:
		name = "char16_t"
		size = 2
	case tpi.SimpleTypeChar32:
		name = "char32_t"
		size = 4
	case tpi.SimpleTypeChar8:
		name = "char8_t"
		size = 1
	case tpi.SimpleTypeSByte:
		name = "int8_t"
		size = 1
	case tpi.SimpleTypeByte:
		name = "uint8_t"
		size = 1
	case tpi.SimpleTypeInt16Short, tpi.SimpleTypeInt16:
		name = "short"
		size = 2
	case tpi.SimpleTypeUInt16Short, tpi.SimpleTypeUInt16:
		name = "unsigned short"
		size = 2
	case tpi.SimpleTypeInt32Long:
		name = "long"
		size = 4
	case tpi.SimpleTypeUInt32Long:
		name = "unsigned long"
		size = 4
	case tpi.SimpleTypeInt32:
		name = "int"
		size = 4
	case tpi.SimpleTypeUInt32:
		name = "unsigned int"
		size = 4
	case tpi.SimpleTypeInt64Quad, tpi.SimpleTypeInt64:
		name = "int64_t"
		size = 8
	case tpi.SimpleTypeUInt64Quad, tpi.SimpleTypeUInt64:
		name = "uint64_t"
		size = 8
	case tpi.SimpleTypeInt128Oct, tpi.SimpleTypeInt128:
		name = "__int128"
		size = 16
	case tpi.SimpleTypeUInt128Oct, tpi.SimpleTypeUInt128:
		name = "unsigned __int128"
		size = 16
	case tpi.SimpleTypeFloat16:
		name = "_Float16"
		size = 2
	case tpi.SimpleTypeFloat32:
		name = "float"
		size = 4
	case tpi.SimpleTypeFloat64:
		name = "double"
		size = 8
	case tpi.SimpleTypeFloat80:
		name = "long double"
		size = 10
	case tpi.SimpleTypeFloat128:
		name = "__float128"
		size = 16
	case tpi.SimpleTypeBool8:
		name = "bool"
		size = 1
	case tpi.SimpleTypeBool16:
		name = "bool16"
		size = 2
	case tpi.SimpleTypeBool32:
		name = "bool32"
		size = 4
	case tpi.SimpleTypeBool64:
		name = "bool64"
		size = 8
	case tpi.SimpleTypeHResult:
		name = "HRESULT"
		size = 4
	default:
		name = "unknown"
		size = 0
	}

	isPointer := mode != tpi.SimpleModeDirect
	if isPointer {
		switch mode {
		case tpi.SimpleModeNearPointer, tpi.SimpleModeNearPointer32:
			size = 4
		case tpi.SimpleModeNearPointer64:
			size = 8
		case tpi.SimpleModeNearPointer128:
			size = 16
		}
	}

	return &PrimitiveType{
		index:     index,
		name:      name,
		size:      size,
		isPointer: isPointer,
	}
}

func (tt *TypeTable) parseTypeRecord(index TypeIndex, record *tpi.TypeRecord) (Type, error) {
	switch record.Kind {
	case tpi.LF_MODIFIER:
		rec, err := tpi.ParseModifierRecord(record.Data)
		if err != nil {
			return nil, err
		}
		return &ModifierType{
			index:        index,
			modifiedType: TypeIndex(rec.ModifiedType),
			isConst:      rec.Modifiers.IsConst(),
			isVolatile:   rec.Modifiers.IsVolatile(),
			isUnaligned:  rec.Modifiers.IsUnaligned(),
		}, nil

	case tpi.LF_POINTER:
		rec, err := tpi.ParsePointerRecord(record.Data)
		if err != nil {
			return nil, err
		}
		mode := rec.Attributes.Mode()
		return &PointerType{
			index:        index,
			referentType: TypeIndex(rec.ReferentType),
			size:         uint64(rec.Attributes.Size()),
			isConst:      rec.Attributes.IsConst(),
			isVolatile:   rec.Attributes.IsVolatile(),
			isReference:  mode == tpi.PointerModeLValueReference,
			isRValue:     mode == tpi.PointerModeRValueReference,
		}, nil

	case tpi.LF_ARRAY:
		rec, err := tpi.ParseArrayRecord(record.Data)
		if err != nil {
			return nil, err
		}
		return &ArrayType{
			index:       index,
			elementType: TypeIndex(rec.ElementType),
			indexType:   TypeIndex(rec.IndexType),
			size:        rec.Size,
			name:        rec.Name,
		}, nil

	case tpi.LF_PROCEDURE:
		rec, err := tpi.ParseProcedureRecord(record.Data)
		if err != nil {
			return nil, err
		}
		return &FunctionType{
			index:          index,
			returnType:     TypeIndex(rec.ReturnType),
			argumentList:   TypeIndex(rec.ArgumentList),
			callingConv:    rec.CallingConv.String(),
			parameterCount: rec.ParameterCount,
		}, nil

	case tpi.LF_MFUNCTION:
		rec, err := tpi.ParseMFunctionRecord(record.Data)
		if err != nil {
			return nil, err
		}
		return &MemberFunctionType{
			index:          index,
			returnType:     TypeIndex(rec.ReturnType),
			classType:      TypeIndex(rec.ClassType),
			thisType:       TypeIndex(rec.ThisType),
			argumentList:   TypeIndex(rec.ArgumentList),
			callingConv:    rec.CallingConv.String(),
			parameterCount: rec.ParameterCount,
			thisAdjust:     rec.ThisAdjust,
		}, nil

	case tpi.LF_CLASS, tpi.LF_CLASS_ST:
		rec, err := tpi.ParseClassRecord(record.Data)
		if err != nil {
			return nil, err
		}
		return &ClassType{
			index:        index,
			name:         rec.Name,
			uniqueName:   rec.UniqueName,
			size:         rec.Size,
			memberCount:  rec.MemberCount,
			fieldList:    TypeIndex(rec.FieldList),
			derivedFrom:  TypeIndex(rec.DerivedFrom),
			vshape:       TypeIndex(rec.VShape),
			isForwardRef: rec.Properties.IsForwardRef(),
		}, nil

	case tpi.LF_STRUCTURE, tpi.LF_STRUCTURE_ST:
		rec, err := tpi.ParseClassRecord(record.Data)
		if err != nil {
			return nil, err
		}
		return &StructType{
			index:        index,
			name:         rec.Name,
			uniqueName:   rec.UniqueName,
			size:         rec.Size,
			memberCount:  rec.MemberCount,
			fieldList:    TypeIndex(rec.FieldList),
			derivedFrom:  TypeIndex(rec.DerivedFrom),
			vshape:       TypeIndex(rec.VShape),
			isForwardRef: rec.Properties.IsForwardRef(),
		}, nil

	case tpi.LF_UNION, tpi.LF_UNION_ST:
		rec, err := tpi.ParseUnionRecord(record.Data)
		if err != nil {
			return nil, err
		}
		return &UnionType{
			index:        index,
			name:         rec.Name,
			uniqueName:   rec.UniqueName,
			size:         rec.Size,
			memberCount:  rec.MemberCount,
			fieldList:    TypeIndex(rec.FieldList),
			isForwardRef: rec.Properties.IsForwardRef(),
		}, nil

	case tpi.LF_ENUM, tpi.LF_ENUM_ST:
		rec, err := tpi.ParseEnumRecord(record.Data)
		if err != nil {
			return nil, err
		}
		return &EnumType{
			index:          index,
			name:           rec.Name,
			uniqueName:     rec.UniqueName,
			underlyingType: TypeIndex(rec.UnderlyingType),
			fieldList:      TypeIndex(rec.FieldList),
			count:          rec.Count,
			isForwardRef:   rec.Properties.IsForwardRef(),
		}, nil

	case tpi.LF_BITFIELD:
		rec, err := tpi.ParseBitFieldRecord(record.Data)
		if err != nil {
			return nil, err
		}
		return &BitfieldType{
			index:          index,
			underlyingType: TypeIndex(rec.Type),
			length:         rec.Length,
			position:       rec.Position,
		}, nil

	default:
		// Return a generic type for unsupported kinds
		return &genericType{
			index: index,
			kind:  TypeKindUnknown,
		}, nil
	}
}

// genericType is used for unsupported type kinds.
type genericType struct {
	index TypeIndex
	kind  TypeKind
}

func (t *genericType) Index() TypeIndex { return t.index }
func (t *genericType) Kind() TypeKind   { return t.kind }
func (t *genericType) Name() string     { return "" }
func (t *genericType) Size() uint64     { return 0 }
