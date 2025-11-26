package main

import (
	"fmt"
	"strings"

	"github.com/skdltmxn/pdb-go/pdb"
	"github.com/spf13/cobra"
)

var (
	symbolsAll      bool
	symbolsKind     string
	symbolsDemangled bool
	symbolsLimit    int
)

var symbolsCmd = &cobra.Command{
	Use:   "symbols <pdb-file>",
	Short: "List symbols in the PDB file",
	Long: `List symbols from a PDB file.

By default, only public symbols are shown. Use --all to include module symbols.
Use --kind to filter by symbol kind (public, function, data, udt, constant).`,
	Args: cobra.ExactArgs(1),
	RunE: runSymbols,
}

func init() {
	symbolsCmd.Flags().BoolVarP(&symbolsAll, "all", "a", false, "show all symbols (including module symbols)")
	symbolsCmd.Flags().StringVarP(&symbolsKind, "kind", "k", "", "filter by symbol kind (public, function, data, udt, constant)")
	symbolsCmd.Flags().BoolVarP(&symbolsDemangled, "demangle", "d", false, "show demangled names")
	symbolsCmd.Flags().IntVarP(&symbolsLimit, "limit", "n", 0, "limit number of symbols shown (0 = unlimited)")
}

func runSymbols(cmd *cobra.Command, args []string) error {
	pdbPath := args[0]

	f, err := pdb.Open(pdbPath)
	if err != nil {
		return fmt.Errorf("failed to open PDB: %w", err)
	}
	defer f.Close()

	symbols, err := f.Symbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	// Determine which kind filter to apply
	var kindFilter pdb.SymbolKind = 0
	hasKindFilter := false
	if symbolsKind != "" {
		hasKindFilter = true
		switch strings.ToLower(symbolsKind) {
		case "public":
			kindFilter = pdb.SymbolKindPublic
		case "function":
			kindFilter = pdb.SymbolKindFunction
		case "data":
			kindFilter = pdb.SymbolKindData
		case "udt":
			kindFilter = pdb.SymbolKindUDT
		case "constant":
			kindFilter = pdb.SymbolKindConstant
		case "local":
			kindFilter = pdb.SymbolKindLocal
		case "parameter":
			kindFilter = pdb.SymbolKindParameter
		default:
			return fmt.Errorf("unknown symbol kind: %s", symbolsKind)
		}
	}

	// Print header
	fmt.Fprintf(output, "%-10s %-8s %-10s %s\n", "KIND", "SECTION", "OFFSET", "NAME")
	fmt.Fprintf(output, "%s\n", strings.Repeat("-", 80))

	count := 0

	if symbolsAll {
		// Iterate all symbols
		for sym := range symbols.All() {
			if hasKindFilter && sym.Kind() != kindFilter {
				continue
			}
			printSymbol(sym)
			count++
			if symbolsLimit > 0 && count >= symbolsLimit {
				break
			}
		}
	} else {
		// Only public symbols
		for sym := range symbols.Public() {
			if hasKindFilter && sym.Kind() != kindFilter {
				continue
			}
			printSymbol(sym)
			count++
			if symbolsLimit > 0 && count >= symbolsLimit {
				break
			}
		}
	}

	fmt.Fprintf(output, "\nTotal: %d symbols\n", count)
	return nil
}

func printSymbol(sym pdb.Symbol) {
	name := sym.Name()
	if symbolsDemangled {
		name = sym.DemangledName()
	}

	section := sym.Section()
	offset := sym.Offset()

	if section == 0 && offset == 0 {
		fmt.Fprintf(output, "%-10s %-8s %-10s %s\n",
			sym.Kind().String(),
			"-",
			"-",
			name)
	} else {
		fmt.Fprintf(output, "%-10s %04X     0x%08X %s\n",
			sym.Kind().String(),
			section,
			offset,
			name)
	}
}
