package main

import (
	"fmt"
	"strings"

	"github.com/skdltmxn/pdb-go/pdb"
	"github.com/spf13/cobra"
)

var (
	typesKind  string
	typesLimit int
)

var typesCmd = &cobra.Command{
	Use:   "types <pdb-file>",
	Short: "List types in the PDB file",
	Long: `List types from a PDB file.

Use --kind to filter by type kind (class, struct, union, enum, function, pointer).`,
	Args: cobra.ExactArgs(1),
	RunE: runTypes,
}

func init() {
	typesCmd.Flags().StringVarP(&typesKind, "kind", "k", "", "filter by type kind (class, struct, union, enum, function, pointer)")
	typesCmd.Flags().IntVarP(&typesLimit, "limit", "n", 0, "limit number of types shown (0 = unlimited)")
}

func runTypes(cmd *cobra.Command, args []string) error {
	pdbPath := args[0]

	f, err := pdb.Open(pdbPath)
	if err != nil {
		return fmt.Errorf("failed to open PDB: %w", err)
	}
	defer f.Close()

	types, err := f.Types()
	if err != nil {
		return fmt.Errorf("failed to get types: %w", err)
	}

	// Determine which kind filter to apply
	var kindFilter pdb.TypeKind = 0
	hasKindFilter := false
	if typesKind != "" {
		hasKindFilter = true
		switch strings.ToLower(typesKind) {
		case "class":
			kindFilter = pdb.TypeKindClass
		case "struct":
			kindFilter = pdb.TypeKindStruct
		case "union":
			kindFilter = pdb.TypeKindUnion
		case "enum":
			kindFilter = pdb.TypeKindEnum
		case "function":
			kindFilter = pdb.TypeKindFunction
		case "pointer":
			kindFilter = pdb.TypeKindPointer
		case "array":
			kindFilter = pdb.TypeKindArray
		case "bitfield":
			kindFilter = pdb.TypeKindBitfield
		default:
			return fmt.Errorf("unknown type kind: %s", typesKind)
		}
	}

	// Print header
	fmt.Fprintf(output, "%-8s %-12s %-8s %s\n", "INDEX", "KIND", "SIZE", "NAME")
	fmt.Fprintf(output, "%s\n", strings.Repeat("-", 80))

	count := 0

	for typ := range types.All() {
		if hasKindFilter && typ.Kind() != kindFilter {
			continue
		}

		printType(typ)
		count++
		if typesLimit > 0 && count >= typesLimit {
			break
		}
	}

	fmt.Fprintf(output, "\nTotal: %d types\n", count)
	return nil
}

func printType(typ pdb.Type) {
	name := typ.Name()
	if name == "" {
		name = "<anonymous>"
	}

	size := typ.Size()
	sizeStr := "-"
	if size > 0 {
		sizeStr = fmt.Sprintf("%d", size)
	}

	fmt.Fprintf(output, "0x%04X   %-12s %-8s %s\n",
		typ.Index(),
		typ.Kind().String(),
		sizeStr,
		name)
}
