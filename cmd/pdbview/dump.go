package main

import (
	"encoding/json"
	"fmt"

	"github.com/skdltmxn/pdb-go/pdb"
	"github.com/spf13/cobra"
)

var (
	dumpFormat string
)

var dumpCmd = &cobra.Command{
	Use:   "dump <pdb-file>",
	Short: "Dump all PDB information",
	Long: `Dump all information from a PDB file in structured format.

Supported formats:
  - text: Human-readable text (default)
  - json: JSON format`,
	Args: cobra.ExactArgs(1),
	RunE: runDump,
}

func init() {
	dumpCmd.Flags().StringVarP(&dumpFormat, "format", "f", "text", "output format (text, json)")
}

func runDump(cmd *cobra.Command, args []string) error {
	pdbPath := args[0]

	f, err := pdb.Open(pdbPath)
	if err != nil {
		return fmt.Errorf("failed to open PDB: %w", err)
	}
	defer f.Close()

	switch dumpFormat {
	case "json":
		return dumpJSON(f, pdbPath)
	case "text":
		return dumpText(f, pdbPath)
	default:
		return fmt.Errorf("unknown format: %s", dumpFormat)
	}
}

type PDBDump struct {
	File    string        `json:"file"`
	Info    *InfoDump     `json:"info"`
	Modules []ModuleDump  `json:"modules"`
	Symbols []SymbolDump  `json:"symbols"`
	Types   []TypeDump    `json:"types"`
}

type InfoDump struct {
	Version    uint32 `json:"version"`
	Signature  uint32 `json:"signature"`
	Age        uint32 `json:"age"`
	GUID       string `json:"guid"`
	BlockSize  uint32 `json:"block_size"`
	NumStreams uint32 `json:"num_streams"`
}

type ModuleDump struct {
	Index          int    `json:"index"`
	Name           string `json:"name"`
	ObjectFileName string `json:"object_file_name"`
	Section        uint16 `json:"section"`
	Offset         int32  `json:"offset"`
	Size           int32  `json:"size"`
}

type SymbolDump struct {
	Name     string `json:"name"`
	Demangled string `json:"demangled"`
	Kind     string `json:"kind"`
	Section  uint16 `json:"section,omitempty"`
	Offset   uint32 `json:"offset,omitempty"`
}

type TypeDump struct {
	Index uint32 `json:"index"`
	Kind  string `json:"kind"`
	Name  string `json:"name,omitempty"`
	Size  uint64 `json:"size,omitempty"`
}

func dumpJSON(f *pdb.File, pdbPath string) error {
	dump := &PDBDump{File: pdbPath}

	// Info
	info, err := f.Info()
	if err == nil {
		dump.Info = &InfoDump{
			Version:   info.Version,
			Signature: info.Signature,
			Age:       info.Age,
			GUID:      formatGUID(info.GUID),
			BlockSize: f.BlockSize(),
		}
		if numStreams, err := f.NumStreams(); err == nil {
			dump.Info.NumStreams = numStreams
		}
	}

	// Modules
	modules, err := f.Modules()
	if err == nil {
		dump.Modules = make([]ModuleDump, len(modules))
		for i, mod := range modules {
			dump.Modules[i] = ModuleDump{
				Index:          mod.Index(),
				Name:           mod.Name(),
				ObjectFileName: mod.ObjectFileName(),
				Section:        mod.Section(),
				Offset:         mod.Offset(),
				Size:           mod.Size(),
			}
		}
	}

	// Symbols (public only for JSON to avoid excessive output)
	symbols, err := f.Symbols()
	if err == nil {
		for sym := range symbols.Public() {
			dump.Symbols = append(dump.Symbols, SymbolDump{
				Name:      sym.Name(),
				Demangled: sym.DemangledName(),
				Kind:      sym.Kind().String(),
				Section:   sym.Section(),
				Offset:    sym.Offset(),
			})
		}
	}

	// Types (named types only for JSON to avoid excessive output)
	types, err := f.Types()
	if err == nil {
		for typ := range types.All() {
			name := typ.Name()
			if name == "" {
				continue // Skip anonymous types in JSON output
			}
			dump.Types = append(dump.Types, TypeDump{
				Index: uint32(typ.Index()),
				Kind:  typ.Kind().String(),
				Name:  name,
				Size:  typ.Size(),
			})
		}
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(dump)
}

func dumpText(f *pdb.File, pdbPath string) error {
	// Reuse the info command
	fmt.Fprintln(output, "=== PDB Information ===")
	if err := runInfo(nil, []string{pdbPath}); err != nil {
		return err
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "=== Modules ===")
	modulesVerbose = true
	if err := runModules(nil, []string{pdbPath}); err != nil {
		return err
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "=== Public Symbols ===")
	symbolsAll = false
	symbolsDemangled = true
	symbolsLimit = 0
	if err := runSymbols(nil, []string{pdbPath}); err != nil {
		return err
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "=== Types ===")
	typesLimit = 0
	if err := runTypes(nil, []string{pdbPath}); err != nil {
		return err
	}

	return nil
}
