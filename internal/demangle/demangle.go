package demangle

import (
	"errors"
	"strings"
)

// Errors
var (
	ErrEmptyInput      = errors.New("demangle: empty input")
	ErrInvalidMangled  = errors.New("demangle: invalid mangled name")
	ErrUnexpectedEnd   = errors.New("demangle: unexpected end of input")
	ErrInvalidBackref  = errors.New("demangle: invalid back-reference")
	ErrUnknownOperator = errors.New("demangle: unknown operator")
	ErrUnknownType     = errors.New("demangle: unknown type")
)

// Demangle converts an MSVC decorated name to readable form.
// If the name is not mangled, it is returned unchanged.
func Demangle(decorated string) (string, error) {
	if len(decorated) == 0 {
		return "", ErrEmptyInput
	}

	// Check if this is a mangled C++ name
	if decorated[0] != '?' {
		// Not a C++ mangled name - might be a C name with underscore prefix
		if len(decorated) > 0 && decorated[0] == '_' {
			return decorated[1:], nil
		}
		return decorated, nil
	}

	// Parse the mangled name
	d := newDemangler(decorated)
	node, err := d.parse()
	if err != nil {
		// On error, return the original name
		return decorated, err
	}

	return node.String(), nil
}

// DemangleToNode parses a mangled name and returns the AST.
func DemangleToNode(decorated string) (Node, error) {
	if len(decorated) == 0 {
		return nil, ErrEmptyInput
	}

	if decorated[0] != '?' {
		return &Identifier{Name: decorated}, nil
	}

	d := newDemangler(decorated)
	return d.parse()
}

// demangler holds parser state.
type demangler struct {
	input string
	pos   int

	// Back-reference tables
	nameBackrefs     [10]string
	nameBackrefCount int
	typeBackrefs     [10]Node
	typeBackrefCount int

	// Template scope tracking
	templateDepth     int
	savedNameBackrefs []backrefState
}

type backrefState struct {
	refs  [10]string
	count int
}

func newDemangler(input string) *demangler {
	return &demangler{
		input: input,
	}
}

func (d *demangler) parse() (Node, error) {
	// Skip leading '?'
	if d.pos < len(d.input) && d.input[d.pos] == '?' {
		d.pos++
	}

	// Check for special intrinsics (??_ patterns)
	if d.peek() == '?' {
		return d.parseSpecialIntrinsic()
	}

	// Parse the name
	name, err := d.parseFullyQualifiedName()
	if err != nil {
		return nil, err
	}

	// After the name comes the encoding
	return d.parseEncoding(name)
}

func (d *demangler) parseSpecialIntrinsic() (Node, error) {
	// Skip second '?'
	d.pos++

	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	switch c {
	case '0':
		// Constructor
		return d.parseCtorDtor(OpConstructor)
	case '1':
		// Destructor
		return d.parseCtorDtor(OpDestructor)
	case '_':
		// Extended special names
		return d.parseExtendedSpecial()
	default:
		// Back up and parse as regular special name
		d.pos--
		return d.parseSpecialName()
	}
}

func (d *demangler) parseCtorDtor(op OperatorKind) (Node, error) {
	name, err := d.parseFullyQualifiedName()
	if err != nil {
		return nil, err
	}

	// Constructor/destructor uses the class name
	var className string
	if len(name.Components) > 0 {
		className = name.Components[len(name.Components)-1].String()
	}

	// Create the operator node
	var opNode Node
	if op == OpDestructor {
		opNode = &Identifier{Name: "~" + className}
	} else {
		opNode = &Identifier{Name: className}
	}

	// Replace the last component with the operator
	name.Components[len(name.Components)-1] = opNode

	return d.parseEncoding(name)
}

func (d *demangler) parseExtendedSpecial() (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	var op OperatorKind
	switch c {
	case '7':
		op = OpVFTable
	case '8':
		op = OpVBTable
	case 'A':
		op = OpTypeof
	case 'B':
		op = OpLocalStaticGuard
	case 'D':
		op = OpVBaseDtor
	case 'E':
		op = OpVectorDeletingDtor
	case 'F':
		op = OpDefaultCtorClosure
	case 'G':
		op = OpScalarDeletingDtor
	case 'H':
		op = OpVectorCtorIterator
	case 'I':
		op = OpVectorDtorIterator
	case 'J':
		op = OpVectorVbaseCtorIterator
	case 'K':
		op = OpVirtualDisplacementMap
	case 'L':
		op = OpEHVectorCtorIterator
	case 'M':
		op = OpEHVectorDtorIterator
	case 'N':
		op = OpEHVectorVbaseCtorIterator
	case 'O':
		op = OpCopyCtorClosure
	case 'R':
		return d.parseRTTI()
	case 'S':
		op = OpLocalVFTable
	case 'T':
		op = OpLocalVFTableCtorClosure
	default:
		d.pos--
		return nil, ErrUnknownOperator
	}

	// Parse the class name
	name, err := d.parseFullyQualifiedName()
	if err != nil {
		return nil, err
	}

	// Append the operator
	name.Components = append(name.Components, &Operator{Op: op})

	return name, nil
}

func (d *demangler) parseRTTI() (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	var op OperatorKind
	switch c {
	case '0':
		op = OpRTTITypeDescriptor
	case '1':
		op = OpRTTIBaseClassDescriptor
	case '2':
		op = OpRTTIBaseClassArray
	case '3':
		op = OpRTTIClassHierarchyDescriptor
	case '4':
		op = OpRTTICompleteObjectLocator
	default:
		d.pos--
		return nil, ErrUnknownOperator
	}

	name, err := d.parseFullyQualifiedName()
	if err != nil {
		return nil, err
	}

	name.Components = append(name.Components, &Operator{Op: op})

	return name, nil
}

func (d *demangler) parseSpecialName() (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	var op OperatorKind
	switch c {
	case '2':
		op = OpNew
	case '3':
		op = OpDelete
	case '4':
		op = OpAssign
	case '5':
		op = OpRightShift
	case '6':
		op = OpLeftShift
	case '7':
		op = OpLogicalNot
	case '8':
		op = OpEqual
	case '9':
		op = OpNotEqual
	case 'A':
		op = OpSubscript
	case 'B':
		// Conversion operator - type follows
		return d.parseConversionOperator()
	case 'C':
		op = OpArrow
	case 'D':
		op = OpDereference
	case 'E':
		op = OpIncrement
	case 'F':
		op = OpDecrement
	case 'G':
		op = OpMinus
	case 'H':
		op = OpPlus
	case 'I':
		op = OpAddressOf
	case 'J':
		op = OpArrowDeref
	case 'K':
		op = OpDivide
	case 'L':
		op = OpModulo
	case 'M':
		op = OpLess
	case 'N':
		op = OpLessEqual
	case 'O':
		op = OpGreater
	case 'P':
		op = OpGreaterEqual
	case 'Q':
		op = OpComma
	case 'R':
		op = OpCall
	case 'S':
		op = OpComplement
	case 'T':
		op = OpXor
	case 'U':
		op = OpBitwiseOr
	case 'V':
		op = OpLogicalAnd
	case 'W':
		op = OpLogicalOr
	case 'X':
		op = OpMultiplyAssign
	case 'Y':
		op = OpPlusAssign
	case 'Z':
		op = OpMinusAssign
	case '_':
		return d.parseExtendedOperator()
	default:
		d.pos--
		return nil, ErrUnknownOperator
	}

	return &Operator{Op: op}, nil
}

func (d *demangler) parseExtendedOperator() (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	var op OperatorKind
	switch c {
	case '0':
		op = OpDivideAssign
	case '1':
		op = OpModuloAssign
	case '2':
		op = OpRightShiftAssign
	case '3':
		op = OpLeftShiftAssign
	case '4':
		op = OpAndAssign
	case '5':
		op = OpOrAssign
	case '6':
		op = OpXorAssign
	case 'U':
		op = OpNewArray
	case 'V':
		op = OpDeleteArray
	default:
		d.pos--
		return nil, ErrUnknownOperator
	}

	return &Operator{Op: op}, nil
}

func (d *demangler) parseConversionOperator() (Node, error) {
	targetType, err := d.parseType()
	if err != nil {
		return nil, err
	}
	return &ConversionOperator{TargetType: targetType}, nil
}

func (d *demangler) parseFullyQualifiedName() (*QualifiedName, error) {
	var components []Node

	for d.pos < len(d.input) {
		if d.peek() == '@' {
			d.pos++
			if d.peek() == '@' {
				d.pos++
				break // End of qualified name
			}
			continue
		}

		part, err := d.parseNameFragment()
		if err != nil {
			return nil, err
		}

		components = append(components, part)
	}

	// Reverse to get natural C++ order
	for i, j := 0, len(components)-1; i < j; i, j = i+1, j-1 {
		components[i], components[j] = components[j], components[i]
	}

	return &QualifiedName{Components: components}, nil
}

func (d *demangler) parseNameFragment() (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.peek()

	// Back-reference (digit)
	if c >= '0' && c <= '9' {
		d.pos++
		idx := int(c - '0')
		if idx >= d.nameBackrefCount {
			return nil, ErrInvalidBackref
		}
		return &Identifier{Name: d.nameBackrefs[idx]}, nil
	}

	// Template instantiation (?$)
	if c == '?' && d.pos+1 < len(d.input) && d.input[d.pos+1] == '$' {
		return d.parseTemplateInstantiation()
	}

	// Special name (?)
	if c == '?' {
		d.pos++
		return d.parseSpecialName()
	}

	// Regular identifier
	return d.parseSimpleName()
}

func (d *demangler) parseSimpleName() (Node, error) {
	start := d.pos
	for d.pos < len(d.input) && d.input[d.pos] != '@' {
		d.pos++
	}

	if d.pos == start {
		return nil, ErrInvalidMangled
	}

	name := d.input[start:d.pos]
	d.memorizeString(name)

	return &Identifier{Name: name}, nil
}

func (d *demangler) parseTemplateInstantiation() (Node, error) {
	// Skip ?$
	d.pos += 2

	// Save backref state
	d.pushBackrefScope()
	defer d.popBackrefScope()

	// Parse template name
	nameNode, err := d.parseSimpleName()
	if err != nil {
		return nil, err
	}

	// Skip @
	if d.pos < len(d.input) && d.input[d.pos] == '@' {
		d.pos++
	}

	// Parse template arguments
	var args []Node
	for d.pos < len(d.input) && d.peek() != '@' {
		arg, err := d.parseTemplateArg()
		if err != nil {
			break
		}
		args = append(args, arg)
	}

	// Skip trailing @
	if d.pos < len(d.input) && d.input[d.pos] == '@' {
		d.pos++
	}

	return &TemplateInstantiation{
		Name:      nameNode,
		Arguments: args,
	}, nil
}

func (d *demangler) parseTemplateArg() (Node, error) {
	c := d.peek()

	// Non-type template argument
	if c == '$' {
		d.pos++
		return d.parseTemplateNonTypeArg()
	}

	// Type template argument
	return d.parseType()
}

func (d *demangler) parseTemplateNonTypeArg() (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	switch c {
	case '0':
		// Integer literal
		val, err := d.parseNumber()
		if err != nil {
			return nil, err
		}
		return &IntegerLiteral{Value: val}, nil
	case '1', '2', 'D', 'E', 'F', 'G', 'H', 'I', 'Q', 'S':
		// Various template parameter types - simplified
		return &Identifier{Name: "?"}, nil
	default:
		d.pos--
		return d.parseType()
	}
}

func (d *demangler) parseEncoding(name *QualifiedName) (Node, error) {
	if d.pos >= len(d.input) {
		return name, nil
	}

	c := d.peek()

	// Function encoding
	switch c {
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y':
		return d.parseFunctionEncoding(name)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return d.parseVariableEncoding(name)
	default:
		return name, nil
	}
}

func (d *demangler) parseFunctionEncoding(name *QualifiedName) (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	var access AccessSpecifier
	var isStatic, isVirtual bool

	switch c {
	case 'A', 'B':
		access = AccessPrivate
	case 'C', 'D':
		access = AccessPrivate
		isStatic = true
	case 'E', 'F':
		access = AccessPrivate
		isVirtual = true
	case 'G', 'H':
		// Private static thunk
		access = AccessPrivate
	case 'I', 'J':
		access = AccessProtected
	case 'K', 'L':
		access = AccessProtected
		isStatic = true
	case 'M', 'N':
		access = AccessProtected
		isVirtual = true
	case 'O', 'P':
		// Protected thunk
		access = AccessProtected
	case 'Q', 'R':
		access = AccessPublic
	case 'S', 'T':
		access = AccessPublic
		isStatic = true
	case 'U', 'V':
		access = AccessPublic
		isVirtual = true
	case 'W', 'X':
		// Public thunk
		access = AccessPublic
	case 'Y', 'Z':
		access = AccessNone // Global function
	}

	// Parse function type
	funcType, err := d.parseFunctionType()
	if err != nil {
		// If function type parsing fails, return just the name
		return &FunctionSymbol{
			Name:       name,
			AccessSpec: access,
			IsStatic:   isStatic,
			IsVirtual:  isVirtual,
		}, nil
	}

	return &FunctionSymbol{
		Name:       name,
		Signature:  funcType,
		AccessSpec: access,
		IsStatic:   isStatic,
		IsVirtual:  isVirtual,
	}, nil
}

func (d *demangler) parseVariableEncoding(name *QualifiedName) (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	var access AccessSpecifier
	var isStatic bool

	switch c {
	case '0':
		access = AccessPrivate
		isStatic = true
	case '1':
		access = AccessProtected
		isStatic = true
	case '2':
		access = AccessPublic
		isStatic = true
	case '3':
		access = AccessNone // Global
	case '4':
		access = AccessNone // Function local
	default:
		d.pos--
	}

	// Parse type
	varType, _ := d.parseType()

	return &VariableSymbol{
		Name:       name,
		Type:       varType,
		AccessSpec: access,
		IsStatic:   isStatic,
	}, nil
}

func (d *demangler) parseFunctionType() (*FunctionType, error) {
	ft := &FunctionType{}

	// Parse calling convention
	cc, err := d.parseCallingConvention()
	if err != nil {
		return nil, err
	}
	ft.CallingConv = cc

	// Parse return type (unless constructor/destructor marked with @)
	if d.peek() != '@' {
		retType, err := d.parseType()
		if err == nil {
			ft.ReturnType = retType
		}
	} else {
		d.pos++ // Skip @
	}

	// Parse parameters
	params, isVariadic, err := d.parseParameters()
	if err == nil {
		ft.Parameters = params
		ft.IsVariadic = isVariadic
	}

	return ft, nil
}

func (d *demangler) parseCallingConvention() (CallingConvention, error) {
	if d.pos >= len(d.input) {
		return CallingConvCdecl, nil
	}

	c := d.consume()

	switch c {
	case 'A', 'B':
		return CallingConvCdecl, nil
	case 'C', 'D':
		return CallingConvPascal, nil
	case 'E', 'F':
		return CallingConvThiscall, nil
	case 'G', 'H':
		return CallingConvStdcall, nil
	case 'I', 'J':
		return CallingConvFastcall, nil
	case 'M', 'N':
		return CallingConvClrcall, nil
	case 'O', 'P':
		return CallingConvEabi, nil
	case 'Q':
		return CallingConvVectorcall, nil
	case 'S':
		return CallingConvSwift, nil
	case 'W':
		return CallingConvSwiftAsync, nil
	default:
		d.pos--
		return CallingConvCdecl, nil
	}
}

func (d *demangler) parseParameters() ([]Node, bool, error) {
	var params []Node
	isVariadic := false

	// Check for void (no parameters)
	if d.peek() == 'X' {
		d.pos++
		return nil, false, nil
	}

	for d.pos < len(d.input) {
		c := d.peek()

		if c == '@' {
			d.pos++
			break
		}

		if c == 'Z' {
			d.pos++
			isVariadic = true
			break
		}

		param, err := d.parseType()
		if err != nil {
			break
		}
		params = append(params, param)
	}

	return params, isVariadic, nil
}

func (d *demangler) parseType() (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.peek()

	// Type back-reference
	if c >= '0' && c <= '9' {
		d.pos++
		idx := int(c - '0')
		if idx >= d.typeBackrefCount {
			return nil, ErrInvalidBackref
		}
		return d.typeBackrefs[idx], nil
	}

	var node Node
	var err error

	switch c {
	// Primitive types
	case 'X':
		d.pos++
		node = &PrimitiveType{Type: PrimVoid}
	case 'C':
		d.pos++
		node = &PrimitiveType{Type: PrimSignedChar}
	case 'D':
		d.pos++
		node = &PrimitiveType{Type: PrimChar}
	case 'E':
		d.pos++
		node = &PrimitiveType{Type: PrimUnsignedChar}
	case 'F':
		d.pos++
		node = &PrimitiveType{Type: PrimShort}
	case 'G':
		d.pos++
		node = &PrimitiveType{Type: PrimUnsignedShort}
	case 'H':
		d.pos++
		node = &PrimitiveType{Type: PrimInt}
	case 'I':
		d.pos++
		node = &PrimitiveType{Type: PrimUnsignedInt}
	case 'J':
		d.pos++
		node = &PrimitiveType{Type: PrimLong}
	case 'K':
		d.pos++
		node = &PrimitiveType{Type: PrimUnsignedLong}
	case 'M':
		d.pos++
		node = &PrimitiveType{Type: PrimFloat}
	case 'N':
		d.pos++
		node = &PrimitiveType{Type: PrimDouble}
	case 'O':
		d.pos++
		node = &PrimitiveType{Type: PrimLongDouble}
	case '_':
		d.pos++
		node, err = d.parseExtendedType()

	// Pointer/reference types
	case 'P':
		d.pos++
		node, err = d.parsePointerType(AffinityPointer)
	case 'Q':
		d.pos++
		node, err = d.parseConstPointerType()
	case 'R':
		d.pos++
		node, err = d.parseVolatilePointerType()
	case 'S':
		d.pos++
		node, err = d.parseConstVolatilePointerType()
	case 'A':
		d.pos++
		node, err = d.parsePointerType(AffinityReference)
	case 'B':
		d.pos++
		node, err = d.parseVolatileReferenceType()

	// $$ extended codes
	case '$':
		d.pos++
		node, err = d.parseDollarType()

	// Tag types
	case 'T':
		d.pos++
		node, err = d.parseTagType(TagUnion)
	case 'U':
		d.pos++
		node, err = d.parseTagType(TagStruct)
	case 'V':
		d.pos++
		node, err = d.parseTagType(TagClass)
	case 'W':
		d.pos++
		node, err = d.parseEnumType()

	// Array type
	case 'Y':
		d.pos++
		node, err = d.parseArrayType()

	default:
		return nil, ErrUnknownType
	}

	if err != nil {
		return nil, err
	}

	// Memorize non-primitive types
	if node != nil && node.Kind() != NodeKindPrimitiveType {
		d.memorizeType(node)
	}

	return node, nil
}

func (d *demangler) parseExtendedType() (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	switch c {
	case 'N':
		return &PrimitiveType{Type: PrimBool}, nil
	case 'J':
		return &PrimitiveType{Type: PrimInt64}, nil
	case 'K':
		return &PrimitiveType{Type: PrimUnsignedInt64}, nil
	case 'W':
		return &PrimitiveType{Type: PrimWChar}, nil
	case 'S':
		return &PrimitiveType{Type: PrimChar16}, nil
	case 'U':
		return &PrimitiveType{Type: PrimChar32}, nil
	default:
		d.pos--
		return nil, ErrUnknownType
	}
}

func (d *demangler) parsePointerType(affinity PointerAffinity) (Node, error) {
	quals := d.parseQualifiers()

	// Check for 64-bit pointer
	is64Bit := false
	if d.peek() == 'E' {
		d.pos++
		is64Bit = true
	}

	pointee, err := d.parseType()
	if err != nil {
		return nil, err
	}

	return &PointerType{
		Pointee:  pointee,
		Affinity: affinity,
		Quals:    quals,
		Is64Bit:  is64Bit,
	}, nil
}

func (d *demangler) parseConstPointerType() (Node, error) {
	quals := d.parseQualifiers()
	quals.IsConst = true

	pointee, err := d.parseType()
	if err != nil {
		return nil, err
	}

	return &PointerType{
		Pointee:  pointee,
		Affinity: AffinityPointer,
		Quals:    quals,
	}, nil
}

func (d *demangler) parseVolatilePointerType() (Node, error) {
	quals := d.parseQualifiers()
	quals.IsVolatile = true

	pointee, err := d.parseType()
	if err != nil {
		return nil, err
	}

	return &PointerType{
		Pointee:  pointee,
		Affinity: AffinityPointer,
		Quals:    quals,
	}, nil
}

func (d *demangler) parseConstVolatilePointerType() (Node, error) {
	quals := d.parseQualifiers()
	quals.IsConst = true
	quals.IsVolatile = true

	pointee, err := d.parseType()
	if err != nil {
		return nil, err
	}

	return &PointerType{
		Pointee:  pointee,
		Affinity: AffinityPointer,
		Quals:    quals,
	}, nil
}

func (d *demangler) parseVolatileReferenceType() (Node, error) {
	quals := d.parseQualifiers()
	quals.IsVolatile = true

	pointee, err := d.parseType()
	if err != nil {
		return nil, err
	}

	return &PointerType{
		Pointee:  pointee,
		Affinity: AffinityReference,
		Quals:    quals,
	}, nil
}

func (d *demangler) parseDollarType() (Node, error) {
	if d.pos >= len(d.input) {
		return nil, ErrUnexpectedEnd
	}

	c := d.consume()

	switch c {
	case '$':
		// $$ codes
		if d.pos < len(d.input) {
			c2 := d.consume()
			switch c2 {
			case 'Q':
				// $$Q - rvalue reference
				return d.parsePointerType(AffinityRValueReference)
			case 'R':
				// $$R - volatile rvalue reference
				quals := d.parseQualifiers()
				quals.IsVolatile = true
				pointee, err := d.parseType()
				if err != nil {
					return nil, err
				}
				return &PointerType{
					Pointee:  pointee,
					Affinity: AffinityRValueReference,
					Quals:    quals,
				}, nil
			case 'A':
				// Function type
				return d.parseFunctionPointerType()
			case 'C':
				// Complex/qualified type
				return d.parseType()
			}
		}
	case 'A':
		// Function
		return d.parseFunctionPointerType()
	case 'Q':
		// Rvalue reference
		return d.parsePointerType(AffinityRValueReference)
	}

	d.pos--
	return nil, ErrUnknownType
}

func (d *demangler) parseFunctionPointerType() (Node, error) {
	// Parse function signature
	return d.parseFunctionType()
}

func (d *demangler) parseQualifiers() Qualifiers {
	var quals Qualifiers

	if d.pos >= len(d.input) {
		return quals
	}

	c := d.peek()

	switch c {
	case 'A':
		d.pos++
		// No qualifiers
	case 'B':
		d.pos++
		quals.IsConst = true
	case 'C':
		d.pos++
		quals.IsVolatile = true
	case 'D':
		d.pos++
		quals.IsConst = true
		quals.IsVolatile = true
	}

	return quals
}

func (d *demangler) parseTagType(tag TagKind) (Node, error) {
	name, err := d.parseFullyQualifiedName()
	if err != nil {
		return nil, err
	}

	return &TagType{
		Tag:  tag,
		Name: name,
	}, nil
}

func (d *demangler) parseEnumType() (Node, error) {
	// Skip underlying type code
	if d.pos < len(d.input) {
		d.pos++
	}

	return d.parseTagType(TagEnum)
}

func (d *demangler) parseArrayType() (Node, error) {
	var dims []uint64

	for d.pos < len(d.input) {
		dim, err := d.parseNumber()
		if err != nil {
			break
		}
		dims = append(dims, uint64(dim))

		if d.peek() != '_' {
			break
		}
		d.pos++
	}

	elemType, err := d.parseType()
	if err != nil {
		return nil, err
	}

	return &ArrayType{
		ElementType: elemType,
		Dimensions:  dims,
	}, nil
}

func (d *demangler) parseNumber() (int64, error) {
	negative := false
	if d.peek() == '?' {
		d.pos++
		negative = true
	}

	c := d.peek()

	// Single digit 1-9
	if c >= '1' && c <= '9' {
		d.pos++
		val := int64(c - '0')
		if negative {
			val = -val
		}
		return val, nil
	}

	// Zero
	if c == '0' {
		d.pos++
		return 0, nil
	}

	// Hex encoded: A-P represent 0-15
	var val int64
	for d.pos < len(d.input) {
		c = d.peek()
		if c == '@' {
			d.pos++
			break
		}

		if c < 'A' || c > 'P' {
			break
		}

		d.pos++
		val = val*16 + int64(c-'A')
	}

	if negative {
		val = -val
	}

	return val, nil
}

// Helper methods

func (d *demangler) peek() byte {
	if d.pos >= len(d.input) {
		return 0
	}
	return d.input[d.pos]
}

func (d *demangler) consume() byte {
	if d.pos >= len(d.input) {
		return 0
	}
	c := d.input[d.pos]
	d.pos++
	return c
}

func (d *demangler) memorizeString(s string) {
	if d.nameBackrefCount < 10 && !d.containsBackref(s) {
		d.nameBackrefs[d.nameBackrefCount] = s
		d.nameBackrefCount++
	}
}

func (d *demangler) containsBackref(s string) bool {
	for i := 0; i < d.nameBackrefCount; i++ {
		if d.nameBackrefs[i] == s {
			return true
		}
	}
	return false
}

func (d *demangler) memorizeType(t Node) {
	if d.typeBackrefCount < 10 {
		d.typeBackrefs[d.typeBackrefCount] = t
		d.typeBackrefCount++
	}
}

func (d *demangler) pushBackrefScope() {
	state := backrefState{
		count: d.nameBackrefCount,
	}
	copy(state.refs[:], d.nameBackrefs[:])
	d.savedNameBackrefs = append(d.savedNameBackrefs, state)

	d.nameBackrefCount = 0
	d.templateDepth++
}

func (d *demangler) popBackrefScope() {
	d.templateDepth--

	if len(d.savedNameBackrefs) > 0 {
		state := d.savedNameBackrefs[len(d.savedNameBackrefs)-1]
		d.savedNameBackrefs = d.savedNameBackrefs[:len(d.savedNameBackrefs)-1]

		d.nameBackrefCount = state.count
		copy(d.nameBackrefs[:], state.refs[:])
	}
}

// DemangleSimple provides a simplified demangling that handles common errors gracefully.
func DemangleSimple(decorated string) string {
	result, err := Demangle(decorated)
	if err != nil {
		return decorated
	}
	return result
}

// IsMangled returns true if the name appears to be an MSVC mangled name.
func IsMangled(name string) bool {
	return len(name) > 0 && (name[0] == '?' || strings.HasPrefix(name, "@?"))
}
