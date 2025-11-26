// Package dbi provides parsing for the DBI (Debug Information) stream.
package dbi

import (
	"errors"
	"fmt"

	"github.com/skdltmxn/pdb-go/internal/stream"
)

// DBI stream version constants
const (
	DBIVersionV41  uint32 = 930803
	DBIVersionV50  uint32 = 19960307
	DBIVersionV60  uint32 = 19970606
	DBIVersionV70  uint32 = 19990903
	DBIVersionV110 uint32 = 20091201
)

// DBI Header size
const DBIHeaderSize = 64

// Machine types
const (
	MachineUnknown   uint16 = 0x0000
	MachineI386      uint16 = 0x014c
	MachineAMD64     uint16 = 0x8664
	MachineARM       uint16 = 0x01c0
	MachineARM64     uint16 = 0xaa64
	MachineARMNT     uint16 = 0x01c4
	MachineIA64      uint16 = 0x0200
)

// Invalid stream index
const InvalidStreamIndex uint16 = 0xFFFF

// Errors
var (
	ErrInvalidDBIHeader   = errors.New("dbi: invalid DBI header")
	ErrUnsupportedVersion = errors.New("dbi: unsupported DBI version")
	ErrTruncatedStream    = errors.New("dbi: truncated stream")
)

// Header represents the DBI stream header.
type Header struct {
	// VersionSignature is always -1
	VersionSignature int32

	// VersionHeader is typically V70 (19990903) or V110 (20091201)
	VersionHeader uint32

	// Age matches the PDB stream Age field
	Age uint32

	// GlobalStreamIndex is the MSF stream index for global symbols
	GlobalStreamIndex uint16

	// BuildNumber encodes toolchain version
	// Bits 0-7: minor version, Bits 8-14: major version, Bit 15: new version flag
	BuildNumber uint16

	// PublicStreamIndex is the MSF stream index for public symbols
	PublicStreamIndex uint16

	// PDBDllVersion is the version of mspdb*.dll
	PDBDllVersion uint16

	// SymRecordStreamIndex is the MSF stream with deduplicated symbol records
	SymRecordStreamIndex uint16

	// PDBDllRbld is the rebuild number of mspdb*.dll
	PDBDllRbld uint16

	// Substream sizes in bytes
	ModInfoSize             uint32
	SectionContributionSize uint32
	SectionMapSize          uint32
	SourceInfoSize          uint32
	TypeServerMapSize       uint32
	MFCTypeServerIndex      uint32
	OptionalDbgHeaderSize   uint32
	ECSubstreamSize         uint32

	// Flags bitfield
	Flags uint16

	// Machine type (e.g., 0x8664 for x86-64)
	Machine uint16

	// Padding
	Padding uint32
}

// BuildMajorVersion returns the major version from BuildNumber.
func (h *Header) BuildMajorVersion() uint16 {
	return (h.BuildNumber >> 8) & 0x7F
}

// BuildMinorVersion returns the minor version from BuildNumber.
func (h *Header) BuildMinorVersion() uint16 {
	return h.BuildNumber & 0xFF
}

// IsIncrementallyLinked returns true if the binary was incrementally linked.
func (h *Header) IsIncrementallyLinked() bool {
	return (h.Flags & 0x01) != 0
}

// IsStripped returns true if private symbols were stripped.
func (h *Header) IsStripped() bool {
	return (h.Flags & 0x02) != 0
}

// HasConflictingTypes returns true if there are conflicting types.
func (h *Header) HasConflictingTypes() bool {
	return (h.Flags & 0x04) != 0
}

// Stream represents a parsed DBI stream.
type Stream struct {
	Header Header

	// Modules is the list of all compilation units
	Modules []ModuleInfo

	// SectionContributions maps addresses to modules
	SectionContributions []SectionContribution

	// SectionMap describes logical segments
	SectionMap *SectionMap

	// SourceFiles contains source file information
	SourceFiles []SourceFileInfo

	// OptionalDbgStreams contains references to additional debug streams
	OptionalDbgStreams *OptionalDbgHeader
}

// ModuleInfo describes a single compilation unit/object file.
type ModuleInfo struct {
	// Opened is unused
	Opened uint32

	// Section contribution for this module
	Section SectionContribution

	// Flags
	Flags uint16

	// ModuleSymStreamIndex is the MSF stream with this module's symbols
	ModuleSymStreamIndex uint16

	// SymByteSize is the size of symbol data in bytes
	SymByteSize uint32

	// C11ByteSize is the size of C11 line info
	C11ByteSize uint32

	// C13ByteSize is the size of C13 line info
	C13ByteSize uint32

	// SourceFileCount is the number of source files
	SourceFileCount uint16

	// SourceFileNameIndex into source file buffer
	SourceFileNameIndex uint32

	// PDBFilePathNameIndex for PDB path
	PDBFilePathNameIndex uint32

	// ModuleName is the object file path
	ModuleName string

	// ObjFileName is the original object file name
	ObjFileName string
}

// SectionContribution describes a module's contribution to a PE section.
type SectionContribution struct {
	Section         uint16
	Padding1        uint16
	Offset          int32
	Size            int32
	Characteristics uint32
	ModuleIndex     uint16
	Padding2        uint16
	DataCrc         uint32
	RelocCrc        uint32
}

// SectionMap contains information about sections.
type SectionMap struct {
	Count     uint16
	LogCount  uint16
	Entries   []SectionMapEntry
}

// SectionMapEntry describes a single section.
type SectionMapEntry struct {
	Flags         uint16
	Ovl           uint16
	Group         uint16
	Frame         uint16
	SectionName   uint16
	ClassName     uint16
	Offset        uint32
	SectionLength uint32
}

// SourceFileInfo contains source file information.
type SourceFileInfo struct {
	ModuleIndex uint16
	FileCount   uint16
	FileOffsets []uint32
	Names       []string
}

// OptionalDbgHeader contains stream indices for additional debug data.
type OptionalDbgHeader struct {
	FPOStreamIndex            uint16
	ExceptionStreamIndex      uint16
	FixupStreamIndex          uint16
	OmapToSrcStreamIndex      uint16
	OmapFromSrcStreamIndex    uint16
	SectionHdrStreamIndex     uint16
	TokenRidMapStreamIndex    uint16
	XDataStreamIndex          uint16
	PDataStreamIndex          uint16
	NewFPOStreamIndex         uint16
	SectionHdrOrigStreamIndex uint16
}

// ParseStream parses a DBI stream from raw data.
func ParseStream(data []byte) (*Stream, error) {
	if len(data) < DBIHeaderSize {
		return nil, ErrInvalidDBIHeader
	}

	r := stream.NewReader(data)
	s := &Stream{}

	// Parse header
	if err := s.parseHeader(r); err != nil {
		return nil, err
	}

	// Parse substreams
	offset := DBIHeaderSize

	// Module Info substream
	if s.Header.ModInfoSize > 0 {
		end := offset + int(s.Header.ModInfoSize)
		if end > len(data) {
			return nil, ErrTruncatedStream
		}
		if err := s.parseModuleInfo(data[offset:end]); err != nil {
			return nil, fmt.Errorf("dbi: failed to parse module info: %w", err)
		}
		offset = end
	}

	// Section Contribution substream
	if s.Header.SectionContributionSize > 0 {
		end := offset + int(s.Header.SectionContributionSize)
		if end > len(data) {
			return nil, ErrTruncatedStream
		}
		if err := s.parseSectionContributions(data[offset:end]); err != nil {
			return nil, fmt.Errorf("dbi: failed to parse section contributions: %w", err)
		}
		offset = end
	}

	// Section Map substream
	if s.Header.SectionMapSize > 0 {
		end := offset + int(s.Header.SectionMapSize)
		if end > len(data) {
			return nil, ErrTruncatedStream
		}
		if err := s.parseSectionMap(data[offset:end]); err != nil {
			return nil, fmt.Errorf("dbi: failed to parse section map: %w", err)
		}
		offset = end
	}

	// Source Info substream
	if s.Header.SourceInfoSize > 0 {
		end := offset + int(s.Header.SourceInfoSize)
		if end > len(data) {
			return nil, ErrTruncatedStream
		}
		// Skip source info parsing for now (complex format)
		offset = end
	}

	// Type Server Map substream
	if s.Header.TypeServerMapSize > 0 {
		offset += int(s.Header.TypeServerMapSize)
	}

	// EC substream
	if s.Header.ECSubstreamSize > 0 {
		offset += int(s.Header.ECSubstreamSize)
	}

	// Optional Debug Header substream
	if s.Header.OptionalDbgHeaderSize > 0 {
		end := offset + int(s.Header.OptionalDbgHeaderSize)
		if end > len(data) {
			return nil, ErrTruncatedStream
		}
		if err := s.parseOptionalDbgHeader(data[offset:end]); err != nil {
			return nil, fmt.Errorf("dbi: failed to parse optional debug header: %w", err)
		}
	}

	return s, nil
}

func (s *Stream) parseHeader(r *stream.Reader) error {
	var err error

	s.Header.VersionSignature, err = r.ReadI32()
	if err != nil {
		return err
	}

	// Version signature should be -1
	if s.Header.VersionSignature != -1 {
		return ErrInvalidDBIHeader
	}

	s.Header.VersionHeader, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.Age, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.GlobalStreamIndex, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.BuildNumber, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.PublicStreamIndex, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.PDBDllVersion, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.SymRecordStreamIndex, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.PDBDllRbld, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.ModInfoSize, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.SectionContributionSize, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.SectionMapSize, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.SourceInfoSize, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.TypeServerMapSize, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.MFCTypeServerIndex, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.OptionalDbgHeaderSize, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.ECSubstreamSize, err = r.ReadU32()
	if err != nil {
		return err
	}

	s.Header.Flags, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.Machine, err = r.ReadU16()
	if err != nil {
		return err
	}

	s.Header.Padding, err = r.ReadU32()
	if err != nil {
		return err
	}

	return nil
}

func (s *Stream) parseModuleInfo(data []byte) error {
	r := stream.NewReader(data)

	for r.Remaining() > 0 {
		var mod ModuleInfo
		var err error

		mod.Opened, err = r.ReadU32()
		if err != nil {
			break
		}

		// Parse section contribution
		mod.Section.Section, err = r.ReadU16()
		if err != nil {
			return err
		}
		mod.Section.Padding1, err = r.ReadU16()
		if err != nil {
			return err
		}
		mod.Section.Offset, err = r.ReadI32()
		if err != nil {
			return err
		}
		mod.Section.Size, err = r.ReadI32()
		if err != nil {
			return err
		}
		mod.Section.Characteristics, err = r.ReadU32()
		if err != nil {
			return err
		}
		mod.Section.ModuleIndex, err = r.ReadU16()
		if err != nil {
			return err
		}
		mod.Section.Padding2, err = r.ReadU16()
		if err != nil {
			return err
		}
		mod.Section.DataCrc, err = r.ReadU32()
		if err != nil {
			return err
		}
		mod.Section.RelocCrc, err = r.ReadU32()
		if err != nil {
			return err
		}

		mod.Flags, err = r.ReadU16()
		if err != nil {
			return err
		}

		mod.ModuleSymStreamIndex, err = r.ReadU16()
		if err != nil {
			return err
		}

		mod.SymByteSize, err = r.ReadU32()
		if err != nil {
			return err
		}

		mod.C11ByteSize, err = r.ReadU32()
		if err != nil {
			return err
		}

		mod.C13ByteSize, err = r.ReadU32()
		if err != nil {
			return err
		}

		mod.SourceFileCount, err = r.ReadU16()
		if err != nil {
			return err
		}

		// Padding
		_, err = r.ReadU16()
		if err != nil {
			return err
		}

		// Unused
		_, err = r.ReadU32()
		if err != nil {
			return err
		}

		mod.SourceFileNameIndex, err = r.ReadU32()
		if err != nil {
			return err
		}

		mod.PDBFilePathNameIndex, err = r.ReadU32()
		if err != nil {
			return err
		}

		// Read module name (null-terminated)
		mod.ModuleName, err = r.ReadCString()
		if err != nil {
			return err
		}

		// Read object file name (null-terminated)
		mod.ObjFileName, err = r.ReadCString()
		if err != nil {
			return err
		}

		// Align to 4-byte boundary
		r.Align(4)

		s.Modules = append(s.Modules, mod)
	}

	return nil
}

// Section Contribution version signatures
const (
	// SectionContribVer60 = 0xeffe0000 + 19970605
	SectionContribVer60 uint32 = 0xF13151F5
	// SectionContribVer2 is an older version format
	SectionContribVer2 uint32 = 0xF12EBA2D
)

func (s *Stream) parseSectionContributions(data []byte) error {
	if len(data) == 0 {
		// Empty section contribution substream is legal
		return nil
	}

	r := stream.NewReader(data)

	// First 4 bytes is version
	version, err := r.ReadU32()
	if err != nil {
		return err
	}

	// Version 60 (0xF13151F5) and Ver2 (0xF12EBA2D) have extra DataCrc and RelocCrc fields
	hasExtraCrc := version == SectionContribVer60 || version == SectionContribVer2

	// Calculate entry size based on version
	entrySize := 28 // Base size: 2+2+4+4+4+2+2 = 20, but actually 28 for V1
	if hasExtraCrc {
		entrySize = 28 + 8 // Extra DataCrc (4) and RelocCrc (4)
	}

	for r.Remaining() >= entrySize {
		var sc SectionContribution

		sc.Section, err = r.ReadU16()
		if err != nil {
			break
		}
		sc.Padding1, err = r.ReadU16()
		if err != nil {
			return err
		}
		sc.Offset, err = r.ReadI32()
		if err != nil {
			return err
		}
		sc.Size, err = r.ReadI32()
		if err != nil {
			return err
		}
		sc.Characteristics, err = r.ReadU32()
		if err != nil {
			return err
		}
		sc.ModuleIndex, err = r.ReadU16()
		if err != nil {
			return err
		}
		sc.Padding2, err = r.ReadU16()
		if err != nil {
			return err
		}

		if hasExtraCrc {
			sc.DataCrc, err = r.ReadU32()
			if err != nil {
				return err
			}
			sc.RelocCrc, err = r.ReadU32()
			if err != nil {
				return err
			}
		}

		s.SectionContributions = append(s.SectionContributions, sc)
	}

	return nil
}

func (s *Stream) parseSectionMap(data []byte) error {
	r := stream.NewReader(data)

	count, err := r.ReadU16()
	if err != nil {
		return err
	}

	logCount, err := r.ReadU16()
	if err != nil {
		return err
	}

	s.SectionMap = &SectionMap{
		Count:    count,
		LogCount: logCount,
		Entries:  make([]SectionMapEntry, 0, count),
	}

	for i := uint16(0); i < count && r.Remaining() >= 20; i++ {
		var entry SectionMapEntry

		entry.Flags, err = r.ReadU16()
		if err != nil {
			return err
		}
		entry.Ovl, err = r.ReadU16()
		if err != nil {
			return err
		}
		entry.Group, err = r.ReadU16()
		if err != nil {
			return err
		}
		entry.Frame, err = r.ReadU16()
		if err != nil {
			return err
		}
		entry.SectionName, err = r.ReadU16()
		if err != nil {
			return err
		}
		entry.ClassName, err = r.ReadU16()
		if err != nil {
			return err
		}
		entry.Offset, err = r.ReadU32()
		if err != nil {
			return err
		}
		entry.SectionLength, err = r.ReadU32()
		if err != nil {
			return err
		}

		s.SectionMap.Entries = append(s.SectionMap.Entries, entry)
	}

	return nil
}

func (s *Stream) parseOptionalDbgHeader(data []byte) error {
	r := stream.NewReader(data)
	s.OptionalDbgStreams = &OptionalDbgHeader{}

	// Each field is a uint16, read as many as we can
	fields := [](*uint16){
		&s.OptionalDbgStreams.FPOStreamIndex,
		&s.OptionalDbgStreams.ExceptionStreamIndex,
		&s.OptionalDbgStreams.FixupStreamIndex,
		&s.OptionalDbgStreams.OmapToSrcStreamIndex,
		&s.OptionalDbgStreams.OmapFromSrcStreamIndex,
		&s.OptionalDbgStreams.SectionHdrStreamIndex,
		&s.OptionalDbgStreams.TokenRidMapStreamIndex,
		&s.OptionalDbgStreams.XDataStreamIndex,
		&s.OptionalDbgStreams.PDataStreamIndex,
		&s.OptionalDbgStreams.NewFPOStreamIndex,
		&s.OptionalDbgStreams.SectionHdrOrigStreamIndex,
	}

	for _, field := range fields {
		if r.Remaining() < 2 {
			break
		}
		val, err := r.ReadU16()
		if err != nil {
			break
		}
		*field = val
	}

	return nil
}

// ModuleCount returns the number of modules.
func (s *Stream) ModuleCount() int {
	return len(s.Modules)
}

// GetModule returns module info by index.
func (s *Stream) GetModule(index int) (*ModuleInfo, error) {
	if index < 0 || index >= len(s.Modules) {
		return nil, fmt.Errorf("dbi: module index out of range: %d", index)
	}
	return &s.Modules[index], nil
}
