# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

pdb-go is a Go library for parsing Microsoft PDB (Program Database) files, designed as a replacement for the DIA SDK. Module path: `github.com/skdltmxn/pdb-go`

The library supports PDB 7.0 format only (BigMsf) and is optimized for large PDB files (~1GB) with fast symbol lookup.

## Build Commands

```bash
go build ./...          # Build all packages
go test ./...           # Run all tests
go test -v ./...        # Run tests with verbose output
go test -run TestName   # Run a specific test
go vet ./...            # Run static analysis
```

Run the CLI viewer:
```bash
go run ./cmd/pdbview info <file.pdb>
go run ./cmd/pdbview symbols <file.pdb>
go run ./cmd/pdbview lookup <file.pdb> <symbol-name>
```

## Architecture

### Package Structure

```
msf/         - MSF (Multi-Stream File) container format parser
pdb/         - Public API (File, SymbolTable, TypeTable, Module)
internal/    - Internal parsing implementations
  stream/    - Low-level binary reader utilities
  dbi/       - DBI (Debug Information) stream parser
  tpi/       - TPI/IPI (Type/ID Program Information) stream parser
  symbols/   - CodeView symbol record parser, GSI/PSI indices
  demangle/  - MSVC symbol name demangler
cmd/pdbview/ - CLI tool (cobra-based)
```

### Data Flow

1. **MSF Layer** (`msf/`): Opens file, reads superblock, parses stream directory. PDB files are block-based containers with numbered streams.

2. **Stream Access**: Well-known streams have fixed indices:
   - Stream 1: PDB Info (GUID, age)
   - Stream 2: TPI (types)
   - Stream 3: DBI (debug info, module list, symbol stream indices)
   - Stream 4: IPI (ID info)

3. **DBI Stream**: Contains header with stream indices for:
   - `SymRecordStreamIndex`: Symbol record data (actual symbols)
   - `PublicStreamIndex`: PSI hash table for public symbols
   - `GlobalStreamIndex`: GSI hash table for global symbols

4. **Symbol Lookup**: Uses lazy-loaded indices:
   - `NameIndex`: Hash-based O(1) lookup by name
   - `AddressIndex`: Binary search O(log n) by section:offset

### Key Design Decisions

- **Lazy Loading**: All streams and indices use `sync.Once` for thread-safe lazy initialization
- **Streaming Iterators**: `Public()` streams symbols on-demand using Go 1.23 `iter.Seq`
- **Memory Efficiency**: Symbol data kept as raw bytes; parsed on-demand
- **Thread Safety**: `File` is safe for concurrent reads after opening

### Symbol Stream vs Hash Table

Important distinction:
- `SymRecordStreamIndex` → Contains actual symbol records (S_PUB32, S_GPROC32, etc.)
- `PublicStreamIndex` → Contains PSI header + hash table + address map (offsets into symbol stream)
- `GlobalStreamIndex` → Contains GSI header + hash table

## Go Version

Requires Go 1.23.0 or later (uses `iter` package for streaming iterators).
