package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFile string
	output     io.Writer
)

var rootCmd = &cobra.Command{
	Use:   "pdbview",
	Short: "PDB file viewer and analyzer",
	Long: `pdbview is a command-line tool for viewing and analyzing
Microsoft PDB (Program Database) files.

It can display symbols, types, modules, and other debug information
stored in PDB files.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			output = f
		} else {
			output = os.Stdout
		}
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if f, ok := output.(*os.File); ok && f != os.Stdout {
			f.Close()
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "write output to file instead of stdout")

	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(symbolsCmd)
	rootCmd.AddCommand(typesCmd)
	rootCmd.AddCommand(modulesCmd)
	rootCmd.AddCommand(lookupCmd)
	rootCmd.AddCommand(dumpCmd)
}
