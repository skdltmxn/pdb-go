package pdb

import (
	"encoding/binary"
	"fmt"
)

// SectionHeader represents a PE section header.
// This matches the IMAGE_SECTION_HEADER structure.
type SectionHeader struct {
	Name                 [8]byte
	VirtualSize          uint32
	VirtualAddress       uint32 // RVA of the section
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLinenumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint32
}

// NameString returns the section name as a string.
func (s *SectionHeader) NameString() string {
	// Find null terminator or use full 8 bytes
	n := 0
	for n < 8 && s.Name[n] != 0 {
		n++
	}
	return string(s.Name[:n])
}

// SectionHeaders provides access to PE section headers stored in PDB.
type SectionHeaders struct {
	sections []SectionHeader
}

// Count returns the number of sections.
func (sh *SectionHeaders) Count() int {
	return len(sh.sections)
}

// Get returns the section header at the given index (0-based).
func (sh *SectionHeaders) Get(index int) (*SectionHeader, error) {
	if index < 0 || index >= len(sh.sections) {
		return nil, fmt.Errorf("pdb: section index out of range: %d", index)
	}
	return &sh.sections[index], nil
}

// All returns all section headers.
func (sh *SectionHeaders) All() []SectionHeader {
	return sh.sections
}

// ToRVA converts a section:offset pair to an RVA (Relative Virtual Address).
// Section numbers are 1-based (as used in PDB symbols).
// Returns 0 if the section number is invalid.
func (sh *SectionHeaders) ToRVA(section uint16, offset uint32) uint32 {
	if section == 0 || int(section) > len(sh.sections) {
		return 0
	}
	return sh.sections[section-1].VirtualAddress + offset
}

// FindSection finds which section contains the given RVA.
// Returns section number (1-based) and offset within the section.
// Returns 0, 0 if the RVA is not within any section.
func (sh *SectionHeaders) FindSection(rva uint32) (section uint16, offset uint32) {
	for i, sec := range sh.sections {
		if rva >= sec.VirtualAddress && rva < sec.VirtualAddress+sec.VirtualSize {
			return uint16(i + 1), rva - sec.VirtualAddress
		}
	}
	return 0, 0
}

// Section header size in bytes
const sectionHeaderSize = 40

func parseSectionHeaders(data []byte) (*SectionHeaders, error) {
	if len(data) < sectionHeaderSize {
		return &SectionHeaders{}, nil
	}

	numSections := len(data) / sectionHeaderSize
	sections := make([]SectionHeader, numSections)

	for i := 0; i < numSections; i++ {
		offset := i * sectionHeaderSize
		sec := &sections[i]

		copy(sec.Name[:], data[offset:offset+8])
		sec.VirtualSize = binary.LittleEndian.Uint32(data[offset+8:])
		sec.VirtualAddress = binary.LittleEndian.Uint32(data[offset+12:])
		sec.SizeOfRawData = binary.LittleEndian.Uint32(data[offset+16:])
		sec.PointerToRawData = binary.LittleEndian.Uint32(data[offset+20:])
		sec.PointerToRelocations = binary.LittleEndian.Uint32(data[offset+24:])
		sec.PointerToLinenumbers = binary.LittleEndian.Uint32(data[offset+28:])
		sec.NumberOfRelocations = binary.LittleEndian.Uint16(data[offset+32:])
		sec.NumberOfLinenumbers = binary.LittleEndian.Uint16(data[offset+34:])
		sec.Characteristics = binary.LittleEndian.Uint32(data[offset+36:])
	}

	return &SectionHeaders{sections: sections}, nil
}

// Sections returns the PE section headers.
func (f *File) Sections() (*SectionHeaders, error) {
	f.sectionHeadersOnce.Do(func() {
		f.sectionHeaders, f.sectionHeadersErr = f.loadSectionHeaders()
	})

	if f.sectionHeadersErr != nil {
		return nil, f.sectionHeadersErr
	}
	return f.sectionHeaders, nil
}

func (f *File) loadSectionHeaders() (*SectionHeaders, error) {
	dbiStream, err := f.getDBI()
	if err != nil {
		return nil, err
	}

	if dbiStream.OptionalDbgStreams == nil {
		return nil, fmt.Errorf("pdb: no optional debug streams")
	}

	streamIndex := dbiStream.OptionalDbgStreams.SectionHdrStreamIndex
	if streamIndex == 0xFFFF {
		return nil, fmt.Errorf("pdb: no section header stream")
	}

	data, err := f.msf.ReadStream(uint32(streamIndex))
	if err != nil {
		return nil, fmt.Errorf("pdb: failed to read section header stream: %w", err)
	}

	return parseSectionHeaders(data)
}
