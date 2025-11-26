// Package tpi provides parsing for TPI (Type Program Information) and IPI (ID Program Information) streams.
package tpi

// TypeIndex is a reference to a type in the TPI or IPI stream.
type TypeIndex uint32

// FirstUserTypeIndex is the first valid user-defined type index.
// Indices below this are simple/primitive types.
const FirstUserTypeIndex TypeIndex = 0x1000

// IsSimpleType returns true if this is a built-in primitive type.
func (ti TypeIndex) IsSimpleType() bool {
	return ti < FirstUserTypeIndex
}

// SimpleKind extracts the simple type kind (bits 0-7).
func (ti TypeIndex) SimpleKind() SimpleTypeKind {
	return SimpleTypeKind(ti & 0xFF)
}

// SimpleMode extracts the simple type mode (bits 8-11).
func (ti TypeIndex) SimpleMode() SimpleTypeMode {
	return SimpleTypeMode((ti >> 8) & 0x0F)
}

// SimpleTypeKind identifies primitive types.
type SimpleTypeKind uint8

const (
	SimpleTypeNone           SimpleTypeKind = 0x00
	SimpleTypeVoid           SimpleTypeKind = 0x03
	SimpleTypeNotTranslated  SimpleTypeKind = 0x07
	SimpleTypeHResult        SimpleTypeKind = 0x08
	SimpleTypeSignedChar     SimpleTypeKind = 0x10
	SimpleTypeUnsignedChar   SimpleTypeKind = 0x20
	SimpleTypeNarrowChar     SimpleTypeKind = 0x70
	SimpleTypeWideChar       SimpleTypeKind = 0x71
	SimpleTypeChar16         SimpleTypeKind = 0x7a
	SimpleTypeChar32         SimpleTypeKind = 0x7b
	SimpleTypeChar8          SimpleTypeKind = 0x7c
	SimpleTypeSByte          SimpleTypeKind = 0x68
	SimpleTypeByte           SimpleTypeKind = 0x69
	SimpleTypeInt16Short     SimpleTypeKind = 0x11
	SimpleTypeUInt16Short    SimpleTypeKind = 0x21
	SimpleTypeInt16          SimpleTypeKind = 0x72
	SimpleTypeUInt16         SimpleTypeKind = 0x73
	SimpleTypeInt32Long      SimpleTypeKind = 0x12
	SimpleTypeUInt32Long     SimpleTypeKind = 0x22
	SimpleTypeInt32          SimpleTypeKind = 0x74
	SimpleTypeUInt32         SimpleTypeKind = 0x75
	SimpleTypeInt64Quad      SimpleTypeKind = 0x13
	SimpleTypeUInt64Quad     SimpleTypeKind = 0x23
	SimpleTypeInt64          SimpleTypeKind = 0x76
	SimpleTypeUInt64         SimpleTypeKind = 0x77
	SimpleTypeInt128Oct      SimpleTypeKind = 0x14
	SimpleTypeUInt128Oct     SimpleTypeKind = 0x24
	SimpleTypeInt128         SimpleTypeKind = 0x78
	SimpleTypeUInt128        SimpleTypeKind = 0x79
	SimpleTypeFloat16        SimpleTypeKind = 0x46
	SimpleTypeFloat32        SimpleTypeKind = 0x40
	SimpleTypeFloat32PP      SimpleTypeKind = 0x45
	SimpleTypeFloat48        SimpleTypeKind = 0x44
	SimpleTypeFloat64        SimpleTypeKind = 0x41
	SimpleTypeFloat80        SimpleTypeKind = 0x42
	SimpleTypeFloat128       SimpleTypeKind = 0x43
	SimpleTypeComplex16      SimpleTypeKind = 0x56
	SimpleTypeComplex32      SimpleTypeKind = 0x50
	SimpleTypeComplex32PP    SimpleTypeKind = 0x55
	SimpleTypeComplex48      SimpleTypeKind = 0x54
	SimpleTypeComplex64      SimpleTypeKind = 0x51
	SimpleTypeComplex80      SimpleTypeKind = 0x52
	SimpleTypeComplex128     SimpleTypeKind = 0x53
	SimpleTypeBool8          SimpleTypeKind = 0x30
	SimpleTypeBool16         SimpleTypeKind = 0x31
	SimpleTypeBool32         SimpleTypeKind = 0x32
	SimpleTypeBool64         SimpleTypeKind = 0x33
	SimpleTypeBool128        SimpleTypeKind = 0x34
)

// SimpleTypeMode identifies pointer modes for simple types.
type SimpleTypeMode uint8

const (
	SimpleModeDirect        SimpleTypeMode = 0x00
	SimpleModeNearPointer   SimpleTypeMode = 0x01
	SimpleModeFarPointer    SimpleTypeMode = 0x02
	SimpleModeHugePointer   SimpleTypeMode = 0x03
	SimpleModeNearPointer32 SimpleTypeMode = 0x04
	SimpleModeFarPointer32  SimpleTypeMode = 0x05
	SimpleModeNearPointer64 SimpleTypeMode = 0x06
	SimpleModeNearPointer128 SimpleTypeMode = 0x07
)

// TypeRecordKind identifies the type of a type record.
type TypeRecordKind uint16

// Type record kinds (LF_*)
const (
	// Leaf types
	LF_MODIFIER     TypeRecordKind = 0x1001
	LF_POINTER      TypeRecordKind = 0x1002
	LF_ARRAY_ST     TypeRecordKind = 0x1003
	LF_CLASS_ST     TypeRecordKind = 0x1004
	LF_STRUCTURE_ST TypeRecordKind = 0x1005
	LF_UNION_ST     TypeRecordKind = 0x1006
	LF_ENUM_ST      TypeRecordKind = 0x1007
	LF_PROCEDURE    TypeRecordKind = 0x1008
	LF_MFUNCTION    TypeRecordKind = 0x1009
	LF_VTSHAPE      TypeRecordKind = 0x000a
	LF_COBOL0       TypeRecordKind = 0x100a
	LF_COBOL1       TypeRecordKind = 0x100b
	LF_BARRAY       TypeRecordKind = 0x100c
	LF_LABEL        TypeRecordKind = 0x000e
	LF_NULL         TypeRecordKind = 0x000f
	LF_NOTTRAN      TypeRecordKind = 0x0010
	LF_DIMARRAY_ST  TypeRecordKind = 0x1016
	LF_VFTPATH      TypeRecordKind = 0x1017
	LF_PRECOMP_ST   TypeRecordKind = 0x1018
	LF_OEM          TypeRecordKind = 0x1019
	LF_ALIAS_ST     TypeRecordKind = 0x101a
	LF_OEM2         TypeRecordKind = 0x101b

	// New records for C++/CLI
	LF_SKIP         TypeRecordKind = 0x1200
	LF_ARGLIST      TypeRecordKind = 0x1201
	LF_DEFARG_ST    TypeRecordKind = 0x1202
	LF_FIELDLIST    TypeRecordKind = 0x1203
	LF_DERIVED      TypeRecordKind = 0x1204
	LF_BITFIELD     TypeRecordKind = 0x1205
	LF_METHODLIST   TypeRecordKind = 0x1206
	LF_DIMCONU      TypeRecordKind = 0x1207
	LF_DIMCONLU     TypeRecordKind = 0x1208
	LF_DIMVARU      TypeRecordKind = 0x1209
	LF_DIMVARLU     TypeRecordKind = 0x120a
	LF_REFSYM       TypeRecordKind = 0x020c

	// New type records
	LF_BCLASS       TypeRecordKind = 0x1400
	LF_VBCLASS      TypeRecordKind = 0x1401
	LF_IVBCLASS     TypeRecordKind = 0x1402
	LF_ENUMERATE_ST TypeRecordKind = 0x0403
	LF_FRIENDFCN_ST TypeRecordKind = 0x1403
	LF_INDEX        TypeRecordKind = 0x1404
	LF_MEMBER_ST    TypeRecordKind = 0x1405
	LF_STMEMBER_ST  TypeRecordKind = 0x1406
	LF_METHOD_ST    TypeRecordKind = 0x1407
	LF_NESTTYPE_ST  TypeRecordKind = 0x1408
	LF_VFUNCTAB     TypeRecordKind = 0x1409
	LF_FRIENDCLS    TypeRecordKind = 0x140a
	LF_ONEMETHOD_ST TypeRecordKind = 0x140b
	LF_VFUNCOFF     TypeRecordKind = 0x140c
	LF_NESTTYPEEX_ST TypeRecordKind = 0x140d
	LF_MEMBERMODIFY_ST TypeRecordKind = 0x140e
	LF_MANAGED_ST   TypeRecordKind = 0x140f

	// Types with string IDs
	LF_TYPESERVER_ST TypeRecordKind = 0x1501
	LF_ENUMERATE    TypeRecordKind = 0x1502
	LF_ARRAY        TypeRecordKind = 0x1503
	LF_CLASS        TypeRecordKind = 0x1504
	LF_STRUCTURE    TypeRecordKind = 0x1505
	LF_UNION        TypeRecordKind = 0x1506
	LF_ENUM         TypeRecordKind = 0x1507
	LF_DIMARRAY     TypeRecordKind = 0x1508
	LF_PRECOMP      TypeRecordKind = 0x1509
	LF_ALIAS        TypeRecordKind = 0x150a
	LF_DEFARG       TypeRecordKind = 0x150b
	LF_FRIENDFCN    TypeRecordKind = 0x150c
	LF_MEMBER       TypeRecordKind = 0x150d
	LF_STMEMBER     TypeRecordKind = 0x150e
	LF_METHOD       TypeRecordKind = 0x150f
	LF_NESTTYPE     TypeRecordKind = 0x1510
	LF_ONEMETHOD    TypeRecordKind = 0x1511
	LF_NESTTYPEEX   TypeRecordKind = 0x1512
	LF_MEMBERMODIFY TypeRecordKind = 0x1513
	LF_MANAGED      TypeRecordKind = 0x1514
	LF_TYPESERVER2  TypeRecordKind = 0x1515

	// String IDs
	LF_STRIDED_ARRAY    TypeRecordKind = 0x1516
	LF_HLSL             TypeRecordKind = 0x1517
	LF_MODIFIER_EX      TypeRecordKind = 0x1518
	LF_INTERFACE        TypeRecordKind = 0x1519
	LF_BINTERFACE       TypeRecordKind = 0x151a
	LF_VECTOR           TypeRecordKind = 0x151b
	LF_MATRIX           TypeRecordKind = 0x151c
	LF_VFTABLE          TypeRecordKind = 0x151d

	// ID records (IPI stream)
	LF_FUNC_ID          TypeRecordKind = 0x1601
	LF_MFUNC_ID         TypeRecordKind = 0x1602
	LF_BUILDINFO        TypeRecordKind = 0x1603
	LF_SUBSTR_LIST      TypeRecordKind = 0x1604
	LF_STRING_ID        TypeRecordKind = 0x1605
	LF_UDT_SRC_LINE     TypeRecordKind = 0x1606
	LF_UDT_MOD_SRC_LINE TypeRecordKind = 0x1607

	// Padding (0xF0-0xFF)
	LF_PAD0  TypeRecordKind = 0x00F0
	LF_PAD1  TypeRecordKind = 0x00F1
	LF_PAD2  TypeRecordKind = 0x00F2
	LF_PAD3  TypeRecordKind = 0x00F3
	LF_PAD4  TypeRecordKind = 0x00F4
	LF_PAD5  TypeRecordKind = 0x00F5
	LF_PAD6  TypeRecordKind = 0x00F6
	LF_PAD7  TypeRecordKind = 0x00F7
	LF_PAD8  TypeRecordKind = 0x00F8
	LF_PAD9  TypeRecordKind = 0x00F9
	LF_PAD10 TypeRecordKind = 0x00FA
	LF_PAD11 TypeRecordKind = 0x00FB
	LF_PAD12 TypeRecordKind = 0x00FC
	LF_PAD13 TypeRecordKind = 0x00FD
	LF_PAD14 TypeRecordKind = 0x00FE
	LF_PAD15 TypeRecordKind = 0x00FF
)

// IsPadding returns true if this is a padding record.
func (k TypeRecordKind) IsPadding() bool {
	return k >= LF_PAD0 && k <= LF_PAD15
}

// CallingConvention represents function calling conventions.
type CallingConvention uint8

const (
	CallingConvNearC        CallingConvention = 0x00
	CallingConvFarC         CallingConvention = 0x01
	CallingConvNearPascal   CallingConvention = 0x02
	CallingConvFarPascal    CallingConvention = 0x03
	CallingConvNearFast     CallingConvention = 0x04
	CallingConvFarFast      CallingConvention = 0x05
	CallingConvSkipped      CallingConvention = 0x06
	CallingConvNearStd      CallingConvention = 0x07
	CallingConvFarStd       CallingConvention = 0x08
	CallingConvNearSys      CallingConvention = 0x09
	CallingConvFarSys       CallingConvention = 0x0a
	CallingConvThisCall     CallingConvention = 0x0b
	CallingConvMipsCall     CallingConvention = 0x0c
	CallingConvGeneric      CallingConvention = 0x0d
	CallingConvAlphaCall    CallingConvention = 0x0e
	CallingConvPpcCall      CallingConvention = 0x0f
	CallingConvSHCall       CallingConvention = 0x10
	CallingConvArmCall      CallingConvention = 0x11
	CallingConvAM33Call     CallingConvention = 0x12
	CallingConvTriCall      CallingConvention = 0x13
	CallingConvSH5Call      CallingConvention = 0x14
	CallingConvM32RCall     CallingConvention = 0x15
	CallingConvClrCall      CallingConvention = 0x16
	CallingConvInline       CallingConvention = 0x17
	CallingConvNearVector   CallingConvention = 0x18
	CallingConvSwift        CallingConvention = 0x19
	CallingConvSwiftAsync   CallingConvention = 0x1a
)

func (cc CallingConvention) String() string {
	switch cc {
	case CallingConvNearC, CallingConvFarC:
		return "__cdecl"
	case CallingConvNearPascal, CallingConvFarPascal:
		return "__pascal"
	case CallingConvNearFast, CallingConvFarFast:
		return "__fastcall"
	case CallingConvNearStd, CallingConvFarStd:
		return "__stdcall"
	case CallingConvThisCall:
		return "__thiscall"
	case CallingConvClrCall:
		return "__clrcall"
	case CallingConvNearVector:
		return "__vectorcall"
	case CallingConvSwift:
		return "__swift"
	case CallingConvSwiftAsync:
		return "__swiftasync"
	default:
		return ""
	}
}

// PointerKind identifies the type of pointer.
type PointerKind uint8

const (
	PointerKindNear16             PointerKind = 0x00
	PointerKindFar16              PointerKind = 0x01
	PointerKindHuge16             PointerKind = 0x02
	PointerKindBasedOnSegment     PointerKind = 0x03
	PointerKindBasedOnValue       PointerKind = 0x04
	PointerKindBasedOnSegmentValue PointerKind = 0x05
	PointerKindBasedOnAddress     PointerKind = 0x06
	PointerKindBasedOnSegmentAddress PointerKind = 0x07
	PointerKindBasedOnType        PointerKind = 0x08
	PointerKindBasedOnSelf        PointerKind = 0x09
	PointerKindNear32             PointerKind = 0x0a
	PointerKindFar32              PointerKind = 0x0b
	PointerKindNear64             PointerKind = 0x0c
)

// PointerMode distinguishes pointer types.
type PointerMode uint8

const (
	PointerModePointer                PointerMode = 0x00
	PointerModeLValueReference        PointerMode = 0x01
	PointerModePointerToDataMember    PointerMode = 0x02
	PointerModePointerToMemberFunction PointerMode = 0x03
	PointerModeRValueReference        PointerMode = 0x04
)

// PointerAttributes is a bitfield for pointer properties.
type PointerAttributes uint32

func (pa PointerAttributes) Kind() PointerKind {
	return PointerKind(pa & 0x1F)
}

func (pa PointerAttributes) Mode() PointerMode {
	return PointerMode((pa >> 5) & 0x07)
}

func (pa PointerAttributes) IsFlat32() bool {
	return (pa & 0x100) != 0
}

func (pa PointerAttributes) IsVolatile() bool {
	return (pa & 0x200) != 0
}

func (pa PointerAttributes) IsConst() bool {
	return (pa & 0x400) != 0
}

func (pa PointerAttributes) IsUnaligned() bool {
	return (pa & 0x800) != 0
}

func (pa PointerAttributes) IsRestrict() bool {
	return (pa & 0x1000) != 0
}

func (pa PointerAttributes) Size() uint8 {
	return uint8((pa >> 13) & 0xFF)
}

func (pa PointerAttributes) IsMocom() bool {
	return (pa & 0x200000) != 0
}

func (pa PointerAttributes) IsLRef() bool {
	return (pa & 0x400000) != 0
}

func (pa PointerAttributes) IsRRef() bool {
	return (pa & 0x800000) != 0
}

// ClassProperties is a bitfield for class/struct properties.
type ClassProperties uint16

func (cp ClassProperties) IsPacked() bool          { return (cp & 0x0001) != 0 }
func (cp ClassProperties) HasCtor() bool           { return (cp & 0x0002) != 0 }
func (cp ClassProperties) HasOverloadedOps() bool  { return (cp & 0x0004) != 0 }
func (cp ClassProperties) IsNested() bool          { return (cp & 0x0008) != 0 }
func (cp ClassProperties) ContainsNested() bool    { return (cp & 0x0010) != 0 }
func (cp ClassProperties) HasOverloadedAssign() bool { return (cp & 0x0020) != 0 }
func (cp ClassProperties) HasCastOperator() bool   { return (cp & 0x0040) != 0 }
func (cp ClassProperties) IsForwardRef() bool      { return (cp & 0x0080) != 0 }
func (cp ClassProperties) IsScoped() bool          { return (cp & 0x0100) != 0 }
func (cp ClassProperties) HasUniqueName() bool     { return (cp & 0x0200) != 0 }
func (cp ClassProperties) IsSealed() bool          { return (cp & 0x0400) != 0 }
func (cp ClassProperties) Hfa() uint8              { return uint8((cp >> 11) & 0x03) }
func (cp ClassProperties) IsIntrinsic() bool       { return (cp & 0x2000) != 0 }
func (cp ClassProperties) Mocom() uint8            { return uint8((cp >> 14) & 0x03) }

// MethodProperties is a bitfield for method properties.
type MethodProperties uint16

func (mp MethodProperties) Access() uint8       { return uint8(mp & 0x03) }
func (mp MethodProperties) IsIntro() bool       { return (mp & 0x04) != 0 }
func (mp MethodProperties) IsPure() bool        { return (mp & 0x08) != 0 }
func (mp MethodProperties) IsNoInherit() bool   { return (mp & 0x10) != 0 }
func (mp MethodProperties) IsNoConstruct() bool { return (mp & 0x20) != 0 }
func (mp MethodProperties) IsCompGenX() bool    { return (mp & 0x40) != 0 }
func (mp MethodProperties) IsSealed() bool      { return (mp & 0x80) != 0 }

// MethodKind identifies the kind of a method.
type MethodKind uint8

const (
	MethodKindVanilla     MethodKind = 0x00
	MethodKindVirtual     MethodKind = 0x01
	MethodKindStatic      MethodKind = 0x02
	MethodKindFriend      MethodKind = 0x03
	MethodKindIntroVirtual MethodKind = 0x04
	MethodKindPureVirtual MethodKind = 0x05
	MethodKindPureIntro   MethodKind = 0x06
)

// MemberAccess identifies member accessibility.
type MemberAccess uint8

const (
	MemberAccessNone      MemberAccess = 0
	MemberAccessPrivate   MemberAccess = 1
	MemberAccessProtected MemberAccess = 2
	MemberAccessPublic    MemberAccess = 3
)

func (ma MemberAccess) String() string {
	switch ma {
	case MemberAccessPrivate:
		return "private"
	case MemberAccessProtected:
		return "protected"
	case MemberAccessPublic:
		return "public"
	default:
		return ""
	}
}

// ModifierOptions is a bitfield for type modifier options.
type ModifierOptions uint16

func (mo ModifierOptions) IsConst() bool    { return (mo & 0x01) != 0 }
func (mo ModifierOptions) IsVolatile() bool { return (mo & 0x02) != 0 }
func (mo ModifierOptions) IsUnaligned() bool { return (mo & 0x04) != 0 }

// FunctionOptions is a bitfield for function options.
type FunctionOptions uint8

func (fo FunctionOptions) IsCxxReturnUDT() bool { return (fo & 0x01) != 0 }
func (fo FunctionOptions) IsConstructor() bool  { return (fo & 0x02) != 0 }
func (fo FunctionOptions) IsCtorVBase() bool    { return (fo & 0x04) != 0 }
