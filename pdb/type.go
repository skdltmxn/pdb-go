package pdb

import (
	"iter"
	"strings"
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

	// Index by name for named types (protected by byNameOnce)
	byName     map[string][]Type
	byNameOnce sync.Once

	// Index for member lookup (lazy-built, protected by memberIndexOnce)
	memberIndex     *memberNameIndex
	memberIndexOnce sync.Once
}

// memberNameIndex provides fast member name lookup.
type memberNameIndex struct {
	// byName maps member name -> list of members
	byName map[string][]*Member
	// byQualifiedName maps "OwnerName::MemberName" -> list of members
	byQualifiedName map[string][]*Member
	// inheritance maps class name -> list of base class names (for inherited member lookup)
	inheritance map[string][]string
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

		for _, typ := range tt.byName[name] {
			if !yield(typ) {
				return
			}
		}
	}
}

func (tt *TypeTable) buildNameIndex() {
	tt.byNameOnce.Do(func() {
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

// Member represents a class/struct member (field).
type Member struct {
	Name       string
	Type       TypeIndex
	Offset     uint64      // Byte offset within the class/struct (0 for static)
	Access     string      // "public", "protected", "private", or ""
	OwnerType  TypeIndex   // The class/struct that contains this member
	OwnerName  string      // Name of the owner class/struct
	IsStatic   bool        // True if this is a static member
}

// MemberSearchResult represents a member found in search.
type MemberSearchResult struct {
	Member
	// Additional context
}

// FindMembers searches for class/struct members by name across all types.
// Supports both simple name ("fieldName") and qualified name ("ClassName::fieldName").
// For qualified names like "Child::member", it also searches inherited members from base classes.
// Uses cached index for O(1) lookup after first call.
func (tt *TypeTable) FindMembers(name string) iter.Seq[*MemberSearchResult] {
	return func(yield func(*MemberSearchResult) bool) {
		tt.buildMemberIndex()

		if tt.memberIndex == nil {
			return
		}

		// Check if it's a qualified name (contains ::)
		if idx := strings.Index(name, "::"); idx > 0 {
			className := name[:idx]
			memberName := name[idx+2:]

			// Track already yielded members to avoid duplicates
			seen := make(map[string]bool)

			// Search the class and all its base classes
			classesToSearch := tt.getInheritanceChain(className)

			for _, class := range classesToSearch {
				qualifiedName := class + "::" + memberName
				for _, m := range tt.memberIndex.byQualifiedName[qualifiedName] {
					// Create unique key for deduplication
					key := m.OwnerName + "::" + m.Name
					if seen[key] {
						continue
					}
					seen[key] = true

					result := &MemberSearchResult{Member: *m}
					if !yield(result) {
						return
					}
				}
			}
		} else {
			// Simple name search - no inheritance traversal needed
			for _, m := range tt.memberIndex.byName[name] {
				result := &MemberSearchResult{Member: *m}
				if !yield(result) {
					return
				}
			}
		}
	}
}

// getInheritanceChain returns the class and all its ancestor classes (including the class itself).
// Uses BFS to traverse the inheritance hierarchy.
func (tt *TypeTable) getInheritanceChain(className string) []string {
	result := []string{className}
	visited := map[string]bool{className: true}
	queue := []string{className}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Get base classes of current class
		baseClasses := tt.memberIndex.inheritance[current]
		for _, base := range baseClasses {
			if !visited[base] {
				visited[base] = true
				result = append(result, base)
				queue = append(queue, base)
			}
		}
	}

	return result
}

func (tt *TypeTable) buildMemberIndex() {
	tt.memberIndexOnce.Do(func() {
		typeCount := int(tt.tpiStream.TypeCount())

		idx := &memberNameIndex{
			byName:          make(map[string][]*Member, typeCount/4),
			byQualifiedName: make(map[string][]*Member, typeCount/4),
			inheritance:     make(map[string][]string, typeCount/8),
		}

		// Single pass: collect type names and parse class definitions
		// Store parsed class info for deferred inheritance resolution
		type classInfo struct {
			name           string
			fieldListIndex tpi.TypeIndex
			typeIndex      tpi.TypeIndex
		}

		typeNames := make(map[tpi.TypeIndex]string, typeCount/2)
		classes := make([]classInfo, 0, typeCount/4)

		begin := tt.tpiStream.TypeIndexBegin()
		end := tt.tpiStream.TypeIndexEnd()

		for ti := begin; ti < end; ti++ {
			record, err := tt.tpiStream.GetTypeRecord(ti)
			if err != nil || record == nil {
				continue
			}

			switch record.Kind {
			case tpi.LF_CLASS, tpi.LF_CLASS_ST, tpi.LF_STRUCTURE, tpi.LF_STRUCTURE_ST:
				rec, err := tpi.ParseClassRecord(record.Data)
				if err != nil {
					continue
				}
				typeNames[ti] = rec.Name
				if !rec.Properties.IsForwardRef() && rec.FieldList != 0 {
					classes = append(classes, classInfo{
						name:           rec.Name,
						fieldListIndex: rec.FieldList,
						typeIndex:      ti,
					})
				}
			case tpi.LF_UNION, tpi.LF_UNION_ST:
				rec, err := tpi.ParseUnionRecord(record.Data)
				if err != nil {
					continue
				}
				typeNames[ti] = rec.Name
				if !rec.Properties.IsForwardRef() && rec.FieldList != 0 {
					classes = append(classes, classInfo{
						name:           rec.Name,
						fieldListIndex: rec.FieldList,
						typeIndex:      ti,
					})
				}
			}
		}

		// Process collected classes: build member index and inheritance map
		for _, cls := range classes {
			fieldRecord, err := tt.tpiStream.GetTypeRecord(cls.fieldListIndex)
			if err != nil || fieldRecord == nil || fieldRecord.Kind != tpi.LF_FIELDLIST {
				continue
			}

			fieldList, err := tpi.ParseFieldListRecord(fieldRecord.Data)
			if err != nil {
				continue
			}

			for _, member := range fieldList.Members {
				var m *Member

				switch mem := member.(type) {
				case *tpi.MemberRecord:
					m = &Member{
						Name:      mem.Name,
						Type:      TypeIndex(mem.Type),
						Offset:    mem.Offset,
						Access:    tpi.MemberAccess(mem.Access).String(),
						OwnerType: TypeIndex(cls.typeIndex),
						OwnerName: cls.name,
					}
				case *tpi.StaticMemberRecord:
					m = &Member{
						Name:      mem.Name,
						Type:      TypeIndex(mem.Type),
						Offset:    0,
						Access:    tpi.MemberAccess(mem.Access).String(),
						OwnerType: TypeIndex(cls.typeIndex),
						OwnerName: cls.name,
						IsStatic:  true,
					}
				case *tpi.BaseClassRecord:
					if baseName := typeNames[mem.Type]; baseName != "" {
						idx.inheritance[cls.name] = append(idx.inheritance[cls.name], baseName)
					}
				case *tpi.VirtualBaseClassRecord:
					if baseName := typeNames[mem.BaseType]; baseName != "" {
						idx.inheritance[cls.name] = append(idx.inheritance[cls.name], baseName)
					}
				}

				if m != nil {
					idx.byName[m.Name] = append(idx.byName[m.Name], m)
					qualifiedName := cls.name + "::" + m.Name
					idx.byQualifiedName[qualifiedName] = append(idx.byQualifiedName[qualifiedName], m)
				}
			}
		}

		tt.memberIndex = idx
	})
}

// GetMembers returns all members of a class/struct/union type.
func (tt *TypeTable) GetMembers(typeIndex TypeIndex) ([]*Member, error) {
	record, err := tt.tpiStream.GetTypeRecord(tpi.TypeIndex(typeIndex))
	if err != nil || record == nil {
		return nil, ErrTypeNotFound
	}

	var ownerName string
	var fieldListIndex tpi.TypeIndex

	switch record.Kind {
	case tpi.LF_CLASS, tpi.LF_CLASS_ST:
		rec, err := tpi.ParseClassRecord(record.Data)
		if err != nil {
			return nil, err
		}
		ownerName = rec.Name
		fieldListIndex = rec.FieldList
	case tpi.LF_STRUCTURE, tpi.LF_STRUCTURE_ST:
		rec, err := tpi.ParseClassRecord(record.Data)
		if err != nil {
			return nil, err
		}
		ownerName = rec.Name
		fieldListIndex = rec.FieldList
	case tpi.LF_UNION, tpi.LF_UNION_ST:
		rec, err := tpi.ParseUnionRecord(record.Data)
		if err != nil {
			return nil, err
		}
		ownerName = rec.Name
		fieldListIndex = rec.FieldList
	default:
		return nil, ErrTypeNotFound
	}

	if fieldListIndex == 0 {
		return nil, nil
	}

	fieldRecord, err := tt.tpiStream.GetTypeRecord(fieldListIndex)
	if err != nil || fieldRecord == nil || fieldRecord.Kind != tpi.LF_FIELDLIST {
		return nil, nil
	}

	fieldList, err := tpi.ParseFieldListRecord(fieldRecord.Data)
	if err != nil {
		return nil, err
	}

	var members []*Member
	for _, m := range fieldList.Members {
		switch mem := m.(type) {
		case *tpi.MemberRecord:
			members = append(members, &Member{
				Name:      mem.Name,
				Type:      TypeIndex(mem.Type),
				Offset:    mem.Offset,
				Access:    tpi.MemberAccess(mem.Access).String(),
				OwnerType: typeIndex,
				OwnerName: ownerName,
			})
		case *tpi.StaticMemberRecord:
			members = append(members, &Member{
				Name:      mem.Name,
				Type:      TypeIndex(mem.Type),
				Offset:    0,
				Access:    tpi.MemberAccess(mem.Access).String(),
				OwnerType: typeIndex,
				OwnerName: ownerName,
			})
		}
	}

	return members, nil
}
