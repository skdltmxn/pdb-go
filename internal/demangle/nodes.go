// Package demangle provides MSVC C++ name demangling functionality.
package demangle

import (
	"fmt"
	"strings"
)

// NodeKind identifies the type of AST node.
type NodeKind int

const (
	NodeKindUnknown NodeKind = iota
	// Identifier nodes
	NodeKindIdentifier
	NodeKindOperator
	NodeKindConversionOperator
	NodeKindConstructor
	NodeKindDestructor
	NodeKindVTable
	NodeKindVBTable
	NodeKindRTTI
	NodeKindLocalStaticGuard
	NodeKindStringLiteral
	NodeKindTemplateInstantiation
	// Type nodes
	NodeKindPrimitiveType
	NodeKindPointerType
	NodeKindReferenceType
	NodeKindRValueReferenceType
	NodeKindArrayType
	NodeKindFunctionType
	NodeKindMemberPointerType
	NodeKindTagType
	NodeKindQualifiedType
	// Symbol nodes
	NodeKindFunctionSymbol
	NodeKindVariableSymbol
	NodeKindSpecialTableSymbol
	// Container nodes
	NodeKindQualifiedName
	NodeKindTemplateArgs
	NodeKindIntegerLiteral
)

// Node is the interface implemented by all AST nodes.
type Node interface {
	Kind() NodeKind
	fmt.Stringer
}

// QualifiedName represents a fully-qualified C++ name.
type QualifiedName struct {
	Components []Node
}

func (n *QualifiedName) Kind() NodeKind { return NodeKindQualifiedName }

func (n *QualifiedName) String() string {
	if len(n.Components) == 0 {
		return ""
	}

	var parts []string
	for _, c := range n.Components {
		parts = append(parts, c.String())
	}
	return strings.Join(parts, "::")
}

// Identifier represents a simple name.
type Identifier struct {
	Name string
}

func (n *Identifier) Kind() NodeKind { return NodeKindIdentifier }
func (n *Identifier) String() string { return n.Name }

// OperatorKind identifies operator types.
type OperatorKind int

const (
	OpUnknown OperatorKind = iota
	OpConstructor
	OpDestructor
	OpNew
	OpDelete
	OpAssign
	OpRightShift
	OpLeftShift
	OpLogicalNot
	OpEqual
	OpNotEqual
	OpSubscript
	OpConversion
	OpArrow
	OpDereference
	OpIncrement
	OpDecrement
	OpMinus
	OpPlus
	OpAddressOf
	OpArrowDeref
	OpDivide
	OpModulo
	OpLess
	OpLessEqual
	OpGreater
	OpGreaterEqual
	OpComma
	OpCall
	OpComplement
	OpXor
	OpBitwiseOr
	OpLogicalAnd
	OpLogicalOr
	OpMultiplyAssign
	OpPlusAssign
	OpMinusAssign
	OpDivideAssign
	OpModuloAssign
	OpRightShiftAssign
	OpLeftShiftAssign
	OpAndAssign
	OpOrAssign
	OpXorAssign
	OpNewArray
	OpDeleteArray
	OpVFTable
	OpVBTable
	OpVCall
	OpTypeof
	OpLocalStaticGuard
	OpStringLiteral
	OpVBaseDtor
	OpVectorDeletingDtor
	OpDefaultCtorClosure
	OpScalarDeletingDtor
	OpVectorCtorIterator
	OpVectorDtorIterator
	OpVectorVbaseCtorIterator
	OpVirtualDisplacementMap
	OpEHVectorCtorIterator
	OpEHVectorDtorIterator
	OpEHVectorVbaseCtorIterator
	OpCopyCtorClosure
	OpLocalVFTable
	OpLocalVFTableCtorClosure
	OpRTTITypeDescriptor
	OpRTTIBaseClassArray
	OpRTTIBaseClassDescriptor
	OpRTTIClassHierarchyDescriptor
	OpRTTICompleteObjectLocator
	OpSpaceship
	OpCoAwait
)

var operatorNames = map[OperatorKind]string{
	OpNew:             "operator new",
	OpDelete:          "operator delete",
	OpAssign:          "operator=",
	OpRightShift:      "operator>>",
	OpLeftShift:       "operator<<",
	OpLogicalNot:      "operator!",
	OpEqual:           "operator==",
	OpNotEqual:        "operator!=",
	OpSubscript:       "operator[]",
	OpArrow:           "operator->",
	OpDereference:     "operator*",
	OpIncrement:       "operator++",
	OpDecrement:       "operator--",
	OpMinus:           "operator-",
	OpPlus:            "operator+",
	OpAddressOf:       "operator&",
	OpArrowDeref:      "operator->*",
	OpDivide:          "operator/",
	OpModulo:          "operator%",
	OpLess:            "operator<",
	OpLessEqual:       "operator<=",
	OpGreater:         "operator>",
	OpGreaterEqual:    "operator>=",
	OpComma:           "operator,",
	OpCall:            "operator()",
	OpComplement:      "operator~",
	OpXor:             "operator^",
	OpBitwiseOr:       "operator|",
	OpLogicalAnd:      "operator&&",
	OpLogicalOr:       "operator||",
	OpMultiplyAssign:  "operator*=",
	OpPlusAssign:      "operator+=",
	OpMinusAssign:     "operator-=",
	OpDivideAssign:    "operator/=",
	OpModuloAssign:    "operator%=",
	OpRightShiftAssign: "operator>>=",
	OpLeftShiftAssign:  "operator<<=",
	OpAndAssign:       "operator&=",
	OpOrAssign:        "operator|=",
	OpXorAssign:       "operator^=",
	OpNewArray:        "operator new[]",
	OpDeleteArray:     "operator delete[]",
	OpSpaceship:       "operator<=>",
	OpCoAwait:         "operator co_await",
	OpVFTable:         "`vftable'",
	OpVBTable:         "`vbtable'",
	OpVCall:           "`vcall'",
	OpTypeof:          "`typeof'",
	OpLocalStaticGuard: "`local static guard'",
	OpStringLiteral:    "`string'",
	OpVBaseDtor:        "`vbase destructor'",
	OpVectorDeletingDtor: "`vector deleting destructor'",
	OpDefaultCtorClosure: "`default constructor closure'",
	OpScalarDeletingDtor: "`scalar deleting destructor'",
	OpVectorCtorIterator: "`vector constructor iterator'",
	OpVectorDtorIterator: "`vector destructor iterator'",
	OpVectorVbaseCtorIterator: "`vector vbase constructor iterator'",
	OpVirtualDisplacementMap: "`virtual displacement map'",
	OpEHVectorCtorIterator: "`eh vector constructor iterator'",
	OpEHVectorDtorIterator: "`eh vector destructor iterator'",
	OpEHVectorVbaseCtorIterator: "`eh vector vbase constructor iterator'",
	OpCopyCtorClosure: "`copy constructor closure'",
	OpLocalVFTable: "`local vftable'",
	OpLocalVFTableCtorClosure: "`local vftable constructor closure'",
	OpRTTITypeDescriptor: "`RTTI Type Descriptor'",
	OpRTTIBaseClassArray: "`RTTI Base Class Array'",
	OpRTTIBaseClassDescriptor: "`RTTI Base Class Descriptor'",
	OpRTTIClassHierarchyDescriptor: "`RTTI Class Hierarchy Descriptor'",
	OpRTTICompleteObjectLocator: "`RTTI Complete Object Locator'",
}

// Operator represents an operator name.
type Operator struct {
	Op OperatorKind
}

func (n *Operator) Kind() NodeKind { return NodeKindOperator }

func (n *Operator) String() string {
	if name, ok := operatorNames[n.Op]; ok {
		return name
	}
	return "operator?"
}

// ConversionOperator represents a conversion operator.
type ConversionOperator struct {
	TargetType Node
}

func (n *ConversionOperator) Kind() NodeKind { return NodeKindConversionOperator }

func (n *ConversionOperator) String() string {
	return "operator " + n.TargetType.String()
}

// TemplateInstantiation represents a template with arguments.
type TemplateInstantiation struct {
	Name      Node
	Arguments []Node
}

func (n *TemplateInstantiation) Kind() NodeKind { return NodeKindTemplateInstantiation }

func (n *TemplateInstantiation) String() string {
	var args []string
	for _, arg := range n.Arguments {
		args = append(args, arg.String())
	}
	return n.Name.String() + "<" + strings.Join(args, ", ") + ">"
}

// PrimitiveKind identifies primitive types.
type PrimitiveKind int

const (
	PrimVoid PrimitiveKind = iota
	PrimBool
	PrimChar
	PrimSignedChar
	PrimUnsignedChar
	PrimShort
	PrimUnsignedShort
	PrimInt
	PrimUnsignedInt
	PrimLong
	PrimUnsignedLong
	PrimInt64
	PrimUnsignedInt64
	PrimInt128
	PrimUnsignedInt128
	PrimFloat
	PrimDouble
	PrimLongDouble
	PrimWChar
	PrimChar8
	PrimChar16
	PrimChar32
	PrimNullptr
	PrimAuto
	PrimDecltypeAuto
)

var primitiveNames = map[PrimitiveKind]string{
	PrimVoid:           "void",
	PrimBool:           "bool",
	PrimChar:           "char",
	PrimSignedChar:     "signed char",
	PrimUnsignedChar:   "unsigned char",
	PrimShort:          "short",
	PrimUnsignedShort:  "unsigned short",
	PrimInt:            "int",
	PrimUnsignedInt:    "unsigned int",
	PrimLong:           "long",
	PrimUnsignedLong:   "unsigned long",
	PrimInt64:          "__int64",
	PrimUnsignedInt64:  "unsigned __int64",
	PrimInt128:         "__int128",
	PrimUnsignedInt128: "unsigned __int128",
	PrimFloat:          "float",
	PrimDouble:         "double",
	PrimLongDouble:     "long double",
	PrimWChar:          "wchar_t",
	PrimChar8:          "char8_t",
	PrimChar16:         "char16_t",
	PrimChar32:         "char32_t",
	PrimNullptr:        "std::nullptr_t",
	PrimAuto:           "auto",
	PrimDecltypeAuto:   "decltype(auto)",
}

// PrimitiveType represents a fundamental C++ type.
type PrimitiveType struct {
	Type PrimitiveKind
}

func (n *PrimitiveType) Kind() NodeKind { return NodeKindPrimitiveType }

func (n *PrimitiveType) String() string {
	if name, ok := primitiveNames[n.Type]; ok {
		return name
	}
	return "?"
}

// Qualifiers represents CV-qualifiers.
type Qualifiers struct {
	IsConst    bool
	IsVolatile bool
	IsRestrict bool
	IsUnaligned bool
}

func (q Qualifiers) String() string {
	var parts []string
	if q.IsConst {
		parts = append(parts, "const")
	}
	if q.IsVolatile {
		parts = append(parts, "volatile")
	}
	if q.IsRestrict {
		parts = append(parts, "__restrict")
	}
	if q.IsUnaligned {
		parts = append(parts, "__unaligned")
	}
	return strings.Join(parts, " ")
}

func (q Qualifiers) IsEmpty() bool {
	return !q.IsConst && !q.IsVolatile && !q.IsRestrict && !q.IsUnaligned
}

// PointerAffinity distinguishes pointer types.
type PointerAffinity int

const (
	AffinityPointer PointerAffinity = iota
	AffinityReference
	AffinityRValueReference
)

// PointerType represents a pointer, reference, or rvalue reference.
type PointerType struct {
	Pointee  Node
	Affinity PointerAffinity
	Quals    Qualifiers
	Is64Bit  bool
}

func (n *PointerType) Kind() NodeKind {
	switch n.Affinity {
	case AffinityReference:
		return NodeKindReferenceType
	case AffinityRValueReference:
		return NodeKindRValueReferenceType
	default:
		return NodeKindPointerType
	}
}

func (n *PointerType) String() string {
	pointee := n.Pointee.String()

	var suffix string
	switch n.Affinity {
	case AffinityPointer:
		suffix = " *"
	case AffinityReference:
		suffix = " &"
	case AffinityRValueReference:
		suffix = " &&"
	}

	if !n.Quals.IsEmpty() {
		suffix += " " + n.Quals.String()
	}

	return pointee + suffix
}

// ArrayType represents a C++ array type.
type ArrayType struct {
	ElementType Node
	Dimensions  []uint64
}

func (n *ArrayType) Kind() NodeKind { return NodeKindArrayType }

func (n *ArrayType) String() string {
	result := n.ElementType.String()
	for _, dim := range n.Dimensions {
		result += fmt.Sprintf("[%d]", dim)
	}
	return result
}

// CallingConvention represents function calling conventions.
type CallingConvention int

const (
	CallingConvCdecl CallingConvention = iota
	CallingConvPascal
	CallingConvThiscall
	CallingConvStdcall
	CallingConvFastcall
	CallingConvVectorcall
	CallingConvClrcall
	CallingConvEabi
	CallingConvSwift
	CallingConvSwiftAsync
)

var callingConvNames = map[CallingConvention]string{
	CallingConvCdecl:     "__cdecl",
	CallingConvPascal:    "__pascal",
	CallingConvThiscall:  "__thiscall",
	CallingConvStdcall:   "__stdcall",
	CallingConvFastcall:  "__fastcall",
	CallingConvVectorcall: "__vectorcall",
	CallingConvClrcall:   "__clrcall",
	CallingConvEabi:      "__eabi",
	CallingConvSwift:     "__swiftcall",
	CallingConvSwiftAsync: "__swiftasynccall",
}

// FunctionType represents a function signature.
type FunctionType struct {
	CallingConv CallingConvention
	ReturnType  Node
	Parameters  []Node
	Quals       Qualifiers
	IsVariadic  bool
	RefQualifier RefQualifier
}

func (n *FunctionType) Kind() NodeKind { return NodeKindFunctionType }

func (n *FunctionType) String() string {
	var result strings.Builder

	if n.ReturnType != nil {
		result.WriteString(n.ReturnType.String())
		result.WriteString(" ")
	}

	if n.CallingConv != CallingConvCdecl {
		if name, ok := callingConvNames[n.CallingConv]; ok {
			result.WriteString(name)
			result.WriteString(" ")
		}
	}

	result.WriteString("(")
	for i, param := range n.Parameters {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(param.String())
	}
	if n.IsVariadic {
		if len(n.Parameters) > 0 {
			result.WriteString(", ")
		}
		result.WriteString("...")
	}
	result.WriteString(")")

	if !n.Quals.IsEmpty() {
		result.WriteString(" ")
		result.WriteString(n.Quals.String())
	}

	switch n.RefQualifier {
	case RefQualifierLValue:
		result.WriteString(" &")
	case RefQualifierRValue:
		result.WriteString(" &&")
	}

	return result.String()
}

// RefQualifier for member function reference qualifiers.
type RefQualifier int

const (
	RefQualifierNone RefQualifier = iota
	RefQualifierLValue
	RefQualifierRValue
)

// TagKind identifies class types.
type TagKind int

const (
	TagUnion TagKind = iota
	TagStruct
	TagClass
	TagEnum
)

var tagNames = map[TagKind]string{
	TagUnion:  "union",
	TagStruct: "struct",
	TagClass:  "class",
	TagEnum:   "enum",
}

// TagType represents class, struct, union, or enum types.
type TagType struct {
	Tag  TagKind
	Name *QualifiedName
}

func (n *TagType) Kind() NodeKind { return NodeKindTagType }

func (n *TagType) String() string {
	return tagNames[n.Tag] + " " + n.Name.String()
}

// MemberPointerType represents pointer-to-member.
type MemberPointerType struct {
	ClassType  Node
	MemberType Node
	Quals      Qualifiers
}

func (n *MemberPointerType) Kind() NodeKind { return NodeKindMemberPointerType }

func (n *MemberPointerType) String() string {
	result := n.MemberType.String() + " " + n.ClassType.String() + "::*"
	if !n.Quals.IsEmpty() {
		result += " " + n.Quals.String()
	}
	return result
}

// QualifiedType represents a CV-qualified type.
type QualifiedType struct {
	Type  Node
	Quals Qualifiers
}

func (n *QualifiedType) Kind() NodeKind { return NodeKindQualifiedType }

func (n *QualifiedType) String() string {
	if n.Quals.IsEmpty() {
		return n.Type.String()
	}
	return n.Quals.String() + " " + n.Type.String()
}

// AccessSpecifier identifies member accessibility.
type AccessSpecifier int

const (
	AccessNone AccessSpecifier = iota
	AccessPrivate
	AccessProtected
	AccessPublic
)

var accessNames = map[AccessSpecifier]string{
	AccessPrivate:   "private",
	AccessProtected: "protected",
	AccessPublic:    "public",
}

// FunctionSymbol represents a function definition.
type FunctionSymbol struct {
	Name         *QualifiedName
	Signature    *FunctionType
	AccessSpec   AccessSpecifier
	IsStatic     bool
	IsVirtual    bool
}

func (n *FunctionSymbol) Kind() NodeKind { return NodeKindFunctionSymbol }

func (n *FunctionSymbol) String() string {
	var result strings.Builder

	if n.AccessSpec != AccessNone {
		result.WriteString(accessNames[n.AccessSpec])
		result.WriteString(": ")
	}

	if n.IsStatic {
		result.WriteString("static ")
	}
	if n.IsVirtual {
		result.WriteString("virtual ")
	}

	if n.Signature != nil {
		if n.Signature.ReturnType != nil {
			result.WriteString(n.Signature.ReturnType.String())
			result.WriteString(" ")
		}

		if n.Signature.CallingConv != CallingConvCdecl {
			if name, ok := callingConvNames[n.Signature.CallingConv]; ok {
				result.WriteString(name)
				result.WriteString(" ")
			}
		}
	}

	result.WriteString(n.Name.String())

	if n.Signature != nil {
		result.WriteString("(")
		for i, param := range n.Signature.Parameters {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(param.String())
		}
		if n.Signature.IsVariadic {
			if len(n.Signature.Parameters) > 0 {
				result.WriteString(", ")
			}
			result.WriteString("...")
		}
		result.WriteString(")")

		if !n.Signature.Quals.IsEmpty() {
			result.WriteString(" ")
			result.WriteString(n.Signature.Quals.String())
		}
	}

	return result.String()
}

// VariableSymbol represents a variable definition.
type VariableSymbol struct {
	Name       *QualifiedName
	Type       Node
	AccessSpec AccessSpecifier
	IsStatic   bool
}

func (n *VariableSymbol) Kind() NodeKind { return NodeKindVariableSymbol }

func (n *VariableSymbol) String() string {
	var result strings.Builder

	if n.AccessSpec != AccessNone {
		result.WriteString(accessNames[n.AccessSpec])
		result.WriteString(": ")
	}

	if n.IsStatic {
		result.WriteString("static ")
	}

	if n.Type != nil {
		result.WriteString(n.Type.String())
		result.WriteString(" ")
	}

	result.WriteString(n.Name.String())

	return result.String()
}

// IntegerLiteral represents an integer constant.
type IntegerLiteral struct {
	Value    int64
	Negative bool
}

func (n *IntegerLiteral) Kind() NodeKind { return NodeKindIntegerLiteral }

func (n *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", n.Value)
}
