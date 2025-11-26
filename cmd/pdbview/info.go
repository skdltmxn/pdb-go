package main

import (
	"fmt"

	"github.com/skdltmxn/pdb-go/pdb"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <pdb-file>",
	Short: "Display PDB file information",
	Long:  `Display general information about a PDB file including version, GUID, age, and statistics.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	pdbPath := args[0]

	f, err := pdb.Open(pdbPath)
	if err != nil {
		return fmt.Errorf("failed to open PDB: %w", err)
	}
	defer f.Close()

	info, err := f.Info()
	if err != nil {
		return fmt.Errorf("failed to read PDB info: %w", err)
	}

	fmt.Fprintf(output, "PDB File: %s\n", pdbPath)
	fmt.Fprintf(output, "Version: %d\n", info.Version)
	fmt.Fprintf(output, "Signature: 0x%08X\n", info.Signature)
	fmt.Fprintf(output, "Age: %d\n", info.Age)
	fmt.Fprintf(output, "GUID: %s\n", formatGUID(info.GUID))
	fmt.Fprintf(output, "Block Size: %d\n", f.BlockSize())

	numStreams, err := f.NumStreams()
	if err == nil {
		fmt.Fprintf(output, "Number of Streams: %d\n", numStreams)
	}

	moduleCount, err := f.ModuleCount()
	if err == nil {
		fmt.Fprintf(output, "Number of Modules: %d\n", moduleCount)
	}

	symbols, err := f.Symbols()
	if err == nil {
		fmt.Fprintf(output, "Public Symbols: %d\n", symbols.PublicCount())
	}

	types, err := f.Types()
	if err == nil {
		fmt.Fprintf(output, "Types: %d\n", types.Count())
	}

	return nil
}

func formatGUID(guid [16]byte) string {
	return fmt.Sprintf("{%08X-%04X-%04X-%02X%02X-%02X%02X%02X%02X%02X%02X}",
		uint32(guid[0])|uint32(guid[1])<<8|uint32(guid[2])<<16|uint32(guid[3])<<24,
		uint16(guid[4])|uint16(guid[5])<<8,
		uint16(guid[6])|uint16(guid[7])<<8,
		guid[8], guid[9],
		guid[10], guid[11], guid[12], guid[13], guid[14], guid[15])
}
