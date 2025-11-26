package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/skdltmxn/pdb-go/pdb"
	"github.com/spf13/cobra"
)

var (
	lookupDemangled bool
	lookupShowRVA   bool
)

var lookupCmd = &cobra.Command{
	Use:   "lookup <pdb-file> <query>",
	Short: "Look up symbols, types, or members by name or address",
	Long: `Look up symbols, types, or members in a PDB file.

Query can be:
  - Symbol/member name: lookup file.pdb myFunction
    (searches both symbols and class/struct members)
  - Qualified member: lookup file.pdb MyClass::m_value
    (searches specific class member)
  - Address: lookup file.pdb 0x1234
    (searches for symbols at that offset)
  - Type index: lookup file.pdb type:0x1000
    (looks up type by index)`,
	Args: cobra.ExactArgs(2),
	RunE: runLookup,
}

func init() {
	lookupCmd.Flags().BoolVarP(&lookupDemangled, "demangle", "d", false, "show demangled names")
	lookupCmd.Flags().BoolVarP(&lookupShowRVA, "rva", "r", false, "show RVA (Relative Virtual Address)")
}

func runLookup(cmd *cobra.Command, args []string) error {
	pdbPath := args[0]
	query := args[1]

	f, err := pdb.Open(pdbPath)
	if err != nil {
		return fmt.Errorf("failed to open PDB: %w", err)
	}
	defer f.Close()

	// Check if it's a type lookup
	if strings.HasPrefix(query, "type:") {
		return lookupType(f, strings.TrimPrefix(query, "type:"))
	}

	// Check if it's an address lookup
	if strings.HasPrefix(query, "0x") || strings.HasPrefix(query, "0X") {
		return lookupAddress(f, query)
	}

	// Otherwise, it's a unified name lookup (symbols + members)
	return lookupName(f, query)
}

func lookupName(f *pdb.File, name string) error {
	symbols, err := f.Symbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	types, err := f.Types()
	if err != nil {
		return fmt.Errorf("failed to get types: %w", err)
	}

	var sections *pdb.SectionHeaders
	if lookupShowRVA {
		sections, _ = f.Sections()
	}

	symbolCount := 0
	memberCount := 0

	// Check if this is a qualified name (Class::member)
	isQualified := strings.Contains(name, "::")

	// Search symbols first (only if not a qualified member name)
	if !isQualified {
		for sym := range symbols.ByName(name) {
			printSymbolDetail(sym, sections)
			symbolCount++
		}

		// Also search by demangled name if no exact match
		if symbolCount == 0 {
			for sym := range symbols.All() {
				demangled := sym.DemangledName()
				if strings.Contains(demangled, name) || strings.Contains(sym.Name(), name) {
					printSymbolDetail(sym, sections)
					symbolCount++
				}
			}
		}
	}

	// Search class/struct members
	for result := range types.FindMembers(name) {
		printMemberDetail(result)
		memberCount++
	}

	totalFound := symbolCount + memberCount
	if totalFound == 0 {
		fmt.Fprintf(output, "No results found matching '%s'\n", name)
	} else {
		if symbolCount > 0 && memberCount > 0 {
			fmt.Fprintf(output, "\nFound %d symbol(s) and %d member(s)\n", symbolCount, memberCount)
		} else if symbolCount > 0 {
			fmt.Fprintf(output, "\nFound %d symbol(s)\n", symbolCount)
		} else {
			fmt.Fprintf(output, "\nFound %d member(s)\n", memberCount)
		}
	}

	return nil
}

func lookupAddress(f *pdb.File, addrStr string) error {
	addr, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(addrStr, "0x"), "0X"), 16, 32)
	if err != nil {
		return fmt.Errorf("invalid address: %s", addrStr)
	}

	symbols, err := f.Symbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	var sections *pdb.SectionHeaders
	if lookupShowRVA {
		sections, _ = f.Sections()
	}

	// Search all sections for the address
	found := 0
	for sym := range symbols.Public() {
		if sym.Offset() == uint32(addr) {
			printSymbolDetail(sym, sections)
			found++
		}
	}

	if found == 0 {
		fmt.Fprintf(output, "No symbols found at address 0x%08X\n", addr)
	}

	return nil
}

func lookupType(f *pdb.File, indexStr string) error {
	index, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(indexStr, "0x"), "0X"), 16, 32)
	if err != nil {
		return fmt.Errorf("invalid type index: %s", indexStr)
	}

	types, err := f.Types()
	if err != nil {
		return fmt.Errorf("failed to get types: %w", err)
	}

	typ, err := types.ByIndex(pdb.TypeIndex(index))
	if err != nil {
		return fmt.Errorf("type not found: %w", err)
	}

	printTypeDetail(typ)
	return nil
}

func printSymbolDetail(sym pdb.Symbol, sections *pdb.SectionHeaders) {
	fmt.Fprintf(output, "Symbol:\n")
	fmt.Fprintf(output, "  Name: %s\n", sym.Name())
	fmt.Fprintf(output, "  Demangled: %s\n", sym.DemangledName())
	fmt.Fprintf(output, "  Kind: %s\n", sym.Kind().String())
	if sym.Section() != 0 || sym.Offset() != 0 {
		fmt.Fprintf(output, "  Section: 0x%04X\n", sym.Section())
		fmt.Fprintf(output, "  Offset: 0x%08X\n", sym.Offset())
		if sections != nil {
			rva := sections.ToRVA(sym.Section(), sym.Offset())
			fmt.Fprintf(output, "  RVA: 0x%08X\n", rva)
		}
	}

	// Print type-specific information
	switch s := sym.(type) {
	case *pdb.PublicSymbol:
		fmt.Fprintf(output, "  IsCode: %v\n", s.IsCode())
		fmt.Fprintf(output, "  IsFunction: %v\n", s.IsFunction())
	case *pdb.FunctionSymbol:
		fmt.Fprintf(output, "  Length: %d\n", s.Length())
		fmt.Fprintf(output, "  TypeIndex: 0x%04X\n", s.TypeIndex())
	case *pdb.DataSymbol:
		fmt.Fprintf(output, "  TypeIndex: 0x%04X\n", s.TypeIndex())
	}

	fmt.Fprintln(output)
}

func printTypeDetail(typ pdb.Type) {
	fmt.Fprintf(output, "Type:\n")
	fmt.Fprintf(output, "  Index: 0x%04X\n", typ.Index())
	fmt.Fprintf(output, "  Kind: %s\n", typ.Kind().String())
	if typ.Name() != "" {
		fmt.Fprintf(output, "  Name: %s\n", typ.Name())
	}
	if typ.Size() > 0 {
		fmt.Fprintf(output, "  Size: %d\n", typ.Size())
	}

	// Print type-specific information
	switch t := typ.(type) {
	case *pdb.ClassType:
		fmt.Fprintf(output, "  MemberCount: %d\n", t.MemberCount())
		fmt.Fprintf(output, "  FieldList: 0x%04X\n", t.FieldList())
		if t.UniqueName() != "" {
			fmt.Fprintf(output, "  UniqueName: %s\n", t.UniqueName())
		}
		fmt.Fprintf(output, "  IsForwardRef: %v\n", t.IsForwardRef())
	case *pdb.StructType:
		fmt.Fprintf(output, "  MemberCount: %d\n", t.MemberCount())
		fmt.Fprintf(output, "  FieldList: 0x%04X\n", t.FieldList())
		if t.UniqueName() != "" {
			fmt.Fprintf(output, "  UniqueName: %s\n", t.UniqueName())
		}
		fmt.Fprintf(output, "  IsForwardRef: %v\n", t.IsForwardRef())
	case *pdb.EnumType:
		fmt.Fprintf(output, "  UnderlyingType: 0x%04X\n", t.UnderlyingType())
		fmt.Fprintf(output, "  Count: %d\n", t.Count())
		fmt.Fprintf(output, "  FieldList: 0x%04X\n", t.FieldList())
	case *pdb.PointerType:
		fmt.Fprintf(output, "  ReferentType: 0x%04X\n", t.ReferentType())
		fmt.Fprintf(output, "  IsConst: %v\n", t.IsConst())
		fmt.Fprintf(output, "  IsVolatile: %v\n", t.IsVolatile())
		fmt.Fprintf(output, "  IsReference: %v\n", t.IsReference())
	case *pdb.FunctionType:
		fmt.Fprintf(output, "  ReturnType: 0x%04X\n", t.ReturnType())
		fmt.Fprintf(output, "  ArgumentList: 0x%04X\n", t.ArgumentList())
		fmt.Fprintf(output, "  ParameterCount: %d\n", t.ParameterCount())
		if t.CallingConvention() != "" {
			fmt.Fprintf(output, "  CallingConvention: %s\n", t.CallingConvention())
		}
	case *pdb.ArrayType:
		fmt.Fprintf(output, "  ElementType: 0x%04X\n", t.ElementType())
		fmt.Fprintf(output, "  IndexType: 0x%04X\n", t.IndexType())
	}

	fmt.Fprintln(output)
}

func printMemberDetail(result *pdb.MemberSearchResult) {
	fmt.Fprintf(output, "Symbol:\n")
	fmt.Fprintf(output, "  Name: %s::%s\n", result.OwnerName, result.Name)
	fmt.Fprintf(output, "  Demangled: %s::%s\n", result.OwnerName, result.Name)
	if result.IsStatic {
		fmt.Fprintf(output, "  Kind: static_member\n")
	} else {
		fmt.Fprintf(output, "  Kind: member\n")
	}
	fmt.Fprintf(output, "  Section: -\n")
	fmt.Fprintf(output, "  Offset: 0x%08X (in class)\n", result.Offset)
	fmt.Fprintf(output, "  OwnerType: %s (0x%04X)\n", result.OwnerName, result.OwnerType)
	fmt.Fprintf(output, "  MemberType: 0x%04X\n", result.Type)
	if result.Access != "" {
		fmt.Fprintf(output, "  Access: %s\n", result.Access)
	}
	fmt.Fprintln(output)
}
