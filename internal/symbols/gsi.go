// Package symbols provides parsing for CodeView symbol records.
package symbols

import (
	"sort"

	"github.com/skdltmxn/pdb-go/internal/stream"
)

// GSI (Global Symbol Index) provides hash-based symbol lookup.
// It parses the GSI stream format used by both global and public symbols.
type GSI struct {
	// hashRecords contains offsets into the symbol record stream
	hashRecords []HashRecord
	// buckets maps hash values to hash record indices
	buckets []int32
	// numBuckets is the number of hash buckets
	numBuckets uint32
}

// ParseGSI parses a Global Symbol Index stream.
func ParseGSI(data []byte) (*GSI, error) {
	if len(data) < 16 {
		return nil, ErrUnexpectedEnd
	}

	r := stream.NewReader(data)

	// Read GSI header
	verSig, _ := r.ReadU32()
	verHdr, _ := r.ReadU32()
	hrSize, _ := r.ReadU32()
	bucketSize, _ := r.ReadU32()

	_ = verSig // 0xFFFFFFFF
	_ = verHdr // 0xeffe0000 + 19990810

	// Parse hash records
	numRecords := hrSize / 8 // Each record is 8 bytes
	hashRecords := make([]HashRecord, numRecords)

	for i := uint32(0); i < numRecords; i++ {
		offset, _ := r.ReadU32()
		cref, _ := r.ReadU32()
		hashRecords[i] = HashRecord{
			Offset: offset,
			CRef:   cref,
		}
	}

	// Parse bucket bitmap and entries
	// The bucket data uses a compressed bitmap format
	var buckets []int32
	var numBuckets uint32

	if bucketSize > 0 {
		// First part is a bitmap indicating which buckets are present
		// Then actual bucket values follow
		bitmapWords := (bucketSize + 3) / 4
		if bitmapWords > 0x3FFFF { // Sanity check
			bitmapWords = 0x3FFFF
		}

		// For simplicity, we'll build our own hash table from the records
		// The bucket format is complex (bitmap-compressed)
		numBuckets = 4096 // Default bucket count
		buckets = make([]int32, numBuckets)
		for i := range buckets {
			buckets[i] = -1 // Empty
		}

		// Build hash index from records
		for i, rec := range hashRecords {
			if rec.Offset == 0 {
				continue
			}
			// Hash is stored in high bits or we compute from name later
			bucket := uint32(i) % numBuckets
			if buckets[bucket] == -1 {
				buckets[bucket] = int32(i)
			}
		}
	}

	return &GSI{
		hashRecords: hashRecords,
		buckets:     buckets,
		numBuckets:  numBuckets,
	}, nil
}

// RecordOffsets returns all symbol record offsets in the GSI.
func (g *GSI) RecordOffsets() []uint32 {
	offsets := make([]uint32, 0, len(g.hashRecords))
	for _, rec := range g.hashRecords {
		if rec.Offset > 0 {
			// Offset is stored +1, so subtract 1 to get actual offset
			offsets = append(offsets, rec.Offset-1)
		}
	}
	return offsets
}

// PSI (Public Symbol Index) extends GSI with address-sorted lookup.
type PSI struct {
	*GSI
	header  PSIHeader
	addrMap []uint32 // Sorted offsets into symbol record stream by address
}

// ParsePSI parses a Public Symbol Index stream.
func ParsePSI(data []byte) (*PSI, error) {
	if len(data) < 16 {
		return nil, ErrUnexpectedEnd
	}

	r := stream.NewReader(data)

	// First parse GSI header
	verSig, _ := r.ReadU32()
	verHdr, _ := r.ReadU32()
	hrSize, _ := r.ReadU32()
	bucketSize, _ := r.ReadU32()

	_ = verSig
	_ = verHdr

	// Skip hash records and buckets to get to PSI header
	if err := r.Skip(int(hrSize)); err != nil {
		return nil, err
	}
	if err := r.Skip(int(bucketSize)); err != nil {
		return nil, err
	}

	// Read PSI-specific header
	var header PSIHeader
	var err error

	header.SymHash, err = r.ReadU32()
	if err != nil {
		return nil, err
	}

	header.AddrMapSize, err = r.ReadU32()
	if err != nil {
		return nil, err
	}

	header.NumThunks, err = r.ReadU32()
	if err != nil {
		return nil, err
	}

	header.SizeOfThunk, err = r.ReadU32()
	if err != nil {
		return nil, err
	}

	header.ISectThunkTable, err = r.ReadU16()
	if err != nil {
		return nil, err
	}

	header.Padding, err = r.ReadU16()
	if err != nil {
		return nil, err
	}

	header.OffThunkTable, err = r.ReadU32()
	if err != nil {
		return nil, err
	}

	header.NumSects, err = r.ReadU32()
	if err != nil {
		return nil, err
	}

	// Read address map
	numAddrs := header.AddrMapSize / 4
	addrMap := make([]uint32, 0, numAddrs)
	for i := uint32(0); i < numAddrs; i++ {
		offset, err := r.ReadU32()
		if err != nil {
			break
		}
		addrMap = append(addrMap, offset)
	}

	// Parse GSI part separately for hash records
	gsi, err := ParseGSI(data)
	if err != nil {
		return nil, err
	}

	return &PSI{
		GSI:     gsi,
		header:  header,
		addrMap: addrMap,
	}, nil
}

// AddressMap returns the address-sorted symbol offsets.
// These are offsets into the symbol record stream, sorted by symbol address.
func (p *PSI) AddressMap() []uint32 {
	return p.addrMap
}

// SymbolAddress represents a symbol's location for address lookup.
type SymbolAddress struct {
	Section uint16
	Offset  uint32
	SymOffset uint32 // Offset in symbol record stream
}

// AddressIndex provides fast address-based symbol lookup.
type AddressIndex struct {
	entries []SymbolAddress
}

// NewAddressIndex creates an address index from PSI address map and symbol data.
func NewAddressIndex(addrMap []uint32, symData []byte) *AddressIndex {
	entries := make([]SymbolAddress, 0, len(addrMap))

	for _, symOffset := range addrMap {
		if int(symOffset)+10 > len(symData) {
			continue
		}

		// Parse just enough of the symbol to get section:offset
		rec, _, err := ParseSymbolRecord(symData[symOffset:])
		if err != nil {
			continue
		}

		if rec.Kind != S_PUB32 {
			continue
		}

		// Parse public symbol to get address
		sym, err := ParsePublicSym32(rec.Data)
		if err != nil {
			continue
		}

		entries = append(entries, SymbolAddress{
			Section:   sym.Segment,
			Offset:    sym.Offset,
			SymOffset: symOffset,
		})
	}

	// Sort by section then offset
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Section != entries[j].Section {
			return entries[i].Section < entries[j].Section
		}
		return entries[i].Offset < entries[j].Offset
	})

	return &AddressIndex{entries: entries}
}

// FindByAddress finds the symbol at or before the given address.
// Returns the symbol offset and whether an exact match was found.
func (idx *AddressIndex) FindByAddress(section uint16, offset uint32) (symOffset uint32, exact bool, found bool) {
	if len(idx.entries) == 0 {
		return 0, false, false
	}

	// Binary search for the address
	i := sort.Search(len(idx.entries), func(i int) bool {
		if idx.entries[i].Section != section {
			return idx.entries[i].Section > section
		}
		return idx.entries[i].Offset >= offset
	})

	if i < len(idx.entries) && idx.entries[i].Section == section && idx.entries[i].Offset == offset {
		return idx.entries[i].SymOffset, true, true
	}

	// Return the symbol just before this address (containing symbol)
	if i > 0 {
		prev := idx.entries[i-1]
		if prev.Section == section {
			return prev.SymOffset, false, true
		}
	}

	return 0, false, false
}

// NameIndex provides hash-based symbol name lookup.
type NameIndex struct {
	buckets    [][]nameEntry
	numBuckets uint32
}

type nameEntry struct {
	name      string
	symOffset uint32
}

// NewNameIndex creates a name index from symbol data.
func NewNameIndex(symData []byte) *NameIndex {
	const numBuckets = 4096

	idx := &NameIndex{
		buckets:    make([][]nameEntry, numBuckets),
		numBuckets: numBuckets,
	}

	r := stream.NewReader(symData)
	for r.Remaining() > 4 {
		offset := r.Offset()
		rec, size, err := ParseSymbolRecord(symData[offset:])
		if err != nil {
			break
		}

		name := getSymbolName(rec)
		if name != "" {
			bucket := hashName(name) % numBuckets
			idx.buckets[bucket] = append(idx.buckets[bucket], nameEntry{
				name:      name,
				symOffset: uint32(offset),
			})
		}

		r.Skip(size)
	}

	return idx
}

// FindByName finds symbols with the given name.
// Returns offsets into the symbol record stream.
func (idx *NameIndex) FindByName(name string) []uint32 {
	bucket := hashName(name) % idx.numBuckets
	entries := idx.buckets[bucket]

	var results []uint32
	for _, e := range entries {
		if e.name == name {
			results = append(results, e.symOffset)
		}
	}
	return results
}

// hashName computes hash for symbol name lookup.
// This uses a simple hash; the actual PDB uses a more complex hash.
func hashName(name string) uint32 {
	var hash uint32
	for i := 0; i < len(name); i++ {
		hash = hash*31 + uint32(name[i])
	}
	return hash
}

// getSymbolName extracts name from a symbol record.
func getSymbolName(rec *SymbolRecord) string {
	switch rec.Kind {
	case S_PUB32:
		if sym, err := ParsePublicSym32(rec.Data); err == nil {
			return sym.Name
		}
	case S_GPROC32, S_LPROC32, S_GPROC32_ID, S_LPROC32_ID:
		if sym, err := ParseProcSym(rec.Data); err == nil {
			return sym.Name
		}
	case S_GDATA32, S_LDATA32:
		if sym, err := ParseDataSym(rec.Data); err == nil {
			return sym.Name
		}
	case S_UDT:
		if sym, err := ParseUDTSym(rec.Data); err == nil {
			return sym.Name
		}
	case S_CONSTANT:
		if sym, err := ParseConstantSym(rec.Data); err == nil {
			return sym.Name
		}
	}
	return ""
}
