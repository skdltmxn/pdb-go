package main

import (
	"fmt"
	"strings"

	"github.com/skdltmxn/pdb-go/pdb"
	"github.com/spf13/cobra"
)

var (
	modulesVerbose bool
)

var modulesCmd = &cobra.Command{
	Use:   "modules <pdb-file>",
	Short: "List modules (compilation units) in the PDB file",
	Long:  `List all modules (compilation units/object files) in a PDB file.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runModules,
}

func init() {
	modulesCmd.Flags().BoolVarP(&modulesVerbose, "verbose", "v", false, "show detailed module information")
}

func runModules(cmd *cobra.Command, args []string) error {
	pdbPath := args[0]

	f, err := pdb.Open(pdbPath)
	if err != nil {
		return fmt.Errorf("failed to open PDB: %w", err)
	}
	defer f.Close()

	modules, err := f.Modules()
	if err != nil {
		return fmt.Errorf("failed to get modules: %w", err)
	}

	if modulesVerbose {
		fmt.Fprintf(output, "%-5s %-8s %-10s %-8s %-8s %s\n", "INDEX", "SECTION", "OFFSET", "SIZE", "SYMBOLS", "NAME")
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 100))

		for _, mod := range modules {
			fmt.Fprintf(output, "%-5d %04X     0x%08X %-8d %-8d %s\n",
				mod.Index(),
				mod.Section(),
				mod.Offset(),
				mod.Size(),
				mod.SymbolCount(),
				mod.Name())
			if mod.ObjectFileName() != mod.Name() {
				fmt.Fprintf(output, "      Object: %s\n", mod.ObjectFileName())
			}
		}
	} else {
		fmt.Fprintf(output, "%-5s %s\n", "INDEX", "NAME")
		fmt.Fprintf(output, "%s\n", strings.Repeat("-", 80))

		for _, mod := range modules {
			fmt.Fprintf(output, "%-5d %s\n", mod.Index(), mod.Name())
		}
	}

	fmt.Fprintf(output, "\nTotal: %d modules\n", len(modules))
	return nil
}
