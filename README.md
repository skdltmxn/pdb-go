# pdb-go

A Go library for parsing Microsoft PDB (Program Database) files.

## Features

- Pure Go implementation with no external dependencies (except cobra for CLI)
- Supports PDB 7.0 format (BigMsf)
- Optimized for large PDB files with lazy loading and streaming iterators
- Fast symbol lookup: O(1) by name, O(log n) by address
- Thread-safe for concurrent reads
- MSVC symbol name demangling

## Installation

```bash
go get github.com/skdltmxn/pdb-go
```

Requires Go 1.23.0 or later.

## Usage

### Library

```go
package main

import (
    "fmt"
    "github.com/skdltmxn/pdb-go/pdb"
)

func main() {
    // Open PDB file
    f, err := pdb.Open("example.pdb")
    if err != nil {
        panic(err)
    }
    defer f.Close()

    // Get PDB info
    info, _ := f.Info()
    fmt.Printf("GUID: %x\n", info.GUID)

    // Access symbols
    symbols, _ := f.Symbols()

    // Fast lookup by name
    if sym, found := symbols.FindByName("main"); found {
        fmt.Printf("%s at %d:%x\n", sym.Name(), sym.Section(), sym.Offset())
    }

    // Iterate public symbols (streaming, memory efficient)
    for sym := range symbols.Public() {
        fmt.Printf("%s -> %s\n", sym.Name(), sym.DemangledName())
    }

    // Lookup by address
    if sym, found := symbols.FindSymbolContaining(1, 0x1000); found {
        fmt.Printf("Symbol at address: %s\n", sym.Name())
    }

    // Access types
    types, _ := f.Types()
    for t := range types.All() {
        fmt.Printf("Type %d: %s\n", t.Index(), t.Name())
    }

    // Access modules
    modules, _ := f.Modules()
    for _, mod := range modules {
        fmt.Printf("Module: %s\n", mod.Name())
    }
}
```

### CLI Tool

```bash
# Build the CLI
go build -o pdbview ./cmd/pdbview

# Show PDB info
pdbview info example.pdb

# List symbols
pdbview symbols example.pdb
pdbview symbols --public example.pdb
pdbview symbols --limit 100 example.pdb

# Lookup symbol by name
pdbview lookup example.pdb MyFunction

# List types
pdbview types example.pdb

# List modules
pdbview modules example.pdb

# Dump raw stream data
pdbview dump --stream 3 example.pdb
```

## API Overview

### pdb.File

Main entry point for accessing PDB contents.

| Method | Description |
|--------|-------------|
| `Open(path)` | Open PDB file from path |
| `OpenReader(r, size)` | Open PDB from io.ReaderAt |
| `Info()` | Get PDB metadata (GUID, age, version) |
| `Symbols()` | Get symbol table |
| `Types()` | Get type table |
| `Modules()` | Get list of modules/compilands |

### pdb.SymbolTable

| Method | Description |
|--------|-------------|
| `FindByName(name)` | O(1) lookup by exact name |
| `ByName(name)` | Iterator for all symbols with name |
| `FindSymbolContaining(section, offset)` | O(log n) lookup by address |
| `Public()` | Streaming iterator over public symbols |
| `All()` | Iterator over all symbols |

### pdb.TypeTable

| Method | Description |
|--------|-------------|
| `ByIndex(index)` | Get type by index |
| `All()` | Iterator over all types |
| `Count()` | Number of types |

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    pdb (Public API)                 │
│         File, SymbolTable, TypeTable, Module        │
└─────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│  internal/dbi │ │  internal/tpi │ │internal/symbols│
│  DBI Stream   │ │  TPI/IPI      │ │ Symbol Records│
└───────────────┘ └───────────────┘ └───────────────┘
        │                 │                 │
        └─────────────────┼─────────────────┘
                          ▼
              ┌───────────────────────┐
              │    msf (Container)    │
              │ SuperBlock, Directory │
              │   Stream Access       │
              └───────────────────────┘
```

## Performance

Designed for large PDB files (1GB+):

- **Lazy loading**: Streams and indices are loaded only when accessed
- **Streaming iteration**: `Public()` parses symbols on-demand without loading all into memory
- **Hash-based lookup**: `FindByName()` uses hash table for O(1) average lookup
- **Binary search**: `FindSymbolContaining()` uses sorted address index for O(log n) lookup
- **Cached access**: `PublicCached()` available when repeated iteration is needed

## License

MIT License
