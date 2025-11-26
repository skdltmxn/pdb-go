// Package symbols provides parsing for CodeView symbol records.
package symbols

import "github.com/skdltmxn/pdb-go/internal/tpi"

// SymbolRecordKind identifies the type of a symbol record.
type SymbolRecordKind uint16

// Symbol record kinds (S_*)
const (
	S_COMPILE       SymbolRecordKind = 0x0001
	S_REGISTER_16t  SymbolRecordKind = 0x0002
	S_CONSTANT_16t  SymbolRecordKind = 0x0003
	S_UDT_16t       SymbolRecordKind = 0x0004
	S_SSEARCH       SymbolRecordKind = 0x0005
	S_END           SymbolRecordKind = 0x0006
	S_SKIP          SymbolRecordKind = 0x0007
	S_CVRESERVE     SymbolRecordKind = 0x0008
	S_OBJNAME_ST    SymbolRecordKind = 0x0009
	S_ENDARG        SymbolRecordKind = 0x000a
	S_COBOLUDT_16t  SymbolRecordKind = 0x000b
	S_MANYREG_16t   SymbolRecordKind = 0x000c
	S_RETURN        SymbolRecordKind = 0x000d
	S_ENTRYTHIS     SymbolRecordKind = 0x000e
	S_BPREL16       SymbolRecordKind = 0x0100
	S_LDATA16       SymbolRecordKind = 0x0101
	S_GDATA16       SymbolRecordKind = 0x0102
	S_PUB16         SymbolRecordKind = 0x0103
	S_LPROC16       SymbolRecordKind = 0x0104
	S_GPROC16       SymbolRecordKind = 0x0105
	S_THUNK16       SymbolRecordKind = 0x0106
	S_BLOCK16       SymbolRecordKind = 0x0107
	S_WITH16        SymbolRecordKind = 0x0108
	S_LABEL16       SymbolRecordKind = 0x0109
	S_CEXMODEL16    SymbolRecordKind = 0x010a
	S_VFTABLE16     SymbolRecordKind = 0x010b
	S_REGREL16      SymbolRecordKind = 0x010c
	S_BPREL32_16t   SymbolRecordKind = 0x0200
	S_LDATA32_16t   SymbolRecordKind = 0x0201
	S_GDATA32_16t   SymbolRecordKind = 0x0202
	S_PUB32_16t     SymbolRecordKind = 0x0203
	S_LPROC32_16t   SymbolRecordKind = 0x0204
	S_GPROC32_16t   SymbolRecordKind = 0x0205
	S_THUNK32_ST    SymbolRecordKind = 0x0206
	S_BLOCK32_ST    SymbolRecordKind = 0x0207
	S_WITH32_ST     SymbolRecordKind = 0x0208
	S_LABEL32_ST    SymbolRecordKind = 0x0209
	S_CEXMODEL32    SymbolRecordKind = 0x020a
	S_VFTABLE32_16t SymbolRecordKind = 0x020b
	S_REGREL32_16t  SymbolRecordKind = 0x020c
	S_LTHREAD32_16t SymbolRecordKind = 0x020d
	S_GTHREAD32_16t SymbolRecordKind = 0x020e
	S_SLINK32       SymbolRecordKind = 0x020f
	S_LPROCMIPS_16t SymbolRecordKind = 0x0300
	S_GPROCMIPS_16t SymbolRecordKind = 0x0301

	// Symbol records with new format
	S_PROCREF_ST      SymbolRecordKind = 0x0400
	S_DATAREF_ST      SymbolRecordKind = 0x0401
	S_ALIGN           SymbolRecordKind = 0x0402
	S_LPROCREF_ST     SymbolRecordKind = 0x0403
	S_OEM             SymbolRecordKind = 0x0404
	S_TI16_MAX        SymbolRecordKind = 0x1000
	S_REGISTER_ST     SymbolRecordKind = 0x1001
	S_CONSTANT_ST     SymbolRecordKind = 0x1002
	S_UDT_ST          SymbolRecordKind = 0x1003
	S_COBOLUDT_ST     SymbolRecordKind = 0x1004
	S_MANYREG_ST      SymbolRecordKind = 0x1005
	S_BPREL32_ST      SymbolRecordKind = 0x1006
	S_LDATA32_ST      SymbolRecordKind = 0x1007
	S_GDATA32_ST      SymbolRecordKind = 0x1008
	S_PUB32_ST        SymbolRecordKind = 0x1009
	S_LPROC32_ST      SymbolRecordKind = 0x100a
	S_GPROC32_ST      SymbolRecordKind = 0x100b
	S_VFTABLE32       SymbolRecordKind = 0x100c
	S_REGREL32_ST     SymbolRecordKind = 0x100d
	S_LTHREAD32_ST    SymbolRecordKind = 0x100e
	S_GTHREAD32_ST    SymbolRecordKind = 0x100f
	S_LPROCMIPS_ST    SymbolRecordKind = 0x1010
	S_GPROCMIPS_ST    SymbolRecordKind = 0x1011
	S_FRAMEPROC       SymbolRecordKind = 0x1012
	S_COMPILE2_ST     SymbolRecordKind = 0x1013
	S_MANYREG2_ST     SymbolRecordKind = 0x1014
	S_LPROCIA64_ST    SymbolRecordKind = 0x1015
	S_GPROCIA64_ST    SymbolRecordKind = 0x1016
	S_LOCALSLOT_ST    SymbolRecordKind = 0x1017
	S_PARAMSLOT_ST    SymbolRecordKind = 0x1018
	S_ANNOTATION      SymbolRecordKind = 0x1019
	S_GMANPROC_ST     SymbolRecordKind = 0x101a
	S_LMANPROC_ST     SymbolRecordKind = 0x101b
	S_RESERVED1       SymbolRecordKind = 0x101c
	S_RESERVED2       SymbolRecordKind = 0x101d
	S_RESERVED3       SymbolRecordKind = 0x101e
	S_RESERVED4       SymbolRecordKind = 0x101f
	S_LMANDATA_ST     SymbolRecordKind = 0x1020
	S_GMANDATA_ST     SymbolRecordKind = 0x1021
	S_MANFRAMEREL_ST  SymbolRecordKind = 0x1022
	S_MANREGISTER_ST  SymbolRecordKind = 0x1023
	S_MANSLOT_ST      SymbolRecordKind = 0x1024
	S_MANMANYREG_ST   SymbolRecordKind = 0x1025
	S_MANREGREL_ST    SymbolRecordKind = 0x1026
	S_MANMANYREG2_ST  SymbolRecordKind = 0x1027
	S_UNAMESPACE_ST   SymbolRecordKind = 0x1028
	S_ST_MAX          SymbolRecordKind = 0x1100
	S_OBJNAME         SymbolRecordKind = 0x1101
	S_THUNK32         SymbolRecordKind = 0x1102
	S_BLOCK32         SymbolRecordKind = 0x1103
	S_WITH32          SymbolRecordKind = 0x1104
	S_LABEL32         SymbolRecordKind = 0x1105
	S_REGISTER        SymbolRecordKind = 0x1106
	S_CONSTANT        SymbolRecordKind = 0x1107
	S_UDT             SymbolRecordKind = 0x1108
	S_COBOLUDT        SymbolRecordKind = 0x1109
	S_MANYREG         SymbolRecordKind = 0x110a
	S_BPREL32         SymbolRecordKind = 0x110b
	S_LDATA32         SymbolRecordKind = 0x110c
	S_GDATA32         SymbolRecordKind = 0x110d
	S_PUB32           SymbolRecordKind = 0x110e
	S_LPROC32         SymbolRecordKind = 0x110f
	S_GPROC32         SymbolRecordKind = 0x1110
	S_REGREL32        SymbolRecordKind = 0x1111
	S_LTHREAD32       SymbolRecordKind = 0x1112
	S_GTHREAD32       SymbolRecordKind = 0x1113
	S_LPROCMIPS       SymbolRecordKind = 0x1114
	S_GPROCMIPS       SymbolRecordKind = 0x1115
	S_COMPILE2        SymbolRecordKind = 0x1116
	S_MANYREG2        SymbolRecordKind = 0x1117
	S_LPROCIA64       SymbolRecordKind = 0x1118
	S_GPROCIA64       SymbolRecordKind = 0x1119
	S_LOCALSLOT       SymbolRecordKind = 0x111a
	S_PARAMSLOT       SymbolRecordKind = 0x111b
	S_LMANDATA        SymbolRecordKind = 0x111c
	S_GMANDATA        SymbolRecordKind = 0x111d
	S_MANFRAMEREL     SymbolRecordKind = 0x111e
	S_MANREGISTER     SymbolRecordKind = 0x111f
	S_MANSLOT         SymbolRecordKind = 0x1120
	S_MANMANYREG      SymbolRecordKind = 0x1121
	S_MANREGREL       SymbolRecordKind = 0x1122
	S_MANMANYREG2     SymbolRecordKind = 0x1123
	S_UNAMESPACE      SymbolRecordKind = 0x1124
	S_PROCREF         SymbolRecordKind = 0x1125
	S_DATAREF         SymbolRecordKind = 0x1126
	S_LPROCREF        SymbolRecordKind = 0x1127
	S_ANNOTATIONREF   SymbolRecordKind = 0x1128
	S_TOKENREF        SymbolRecordKind = 0x1129
	S_GMANPROC        SymbolRecordKind = 0x112a
	S_LMANPROC        SymbolRecordKind = 0x112b
	S_TRAMPOLINE      SymbolRecordKind = 0x112c
	S_MANCONSTANT     SymbolRecordKind = 0x112d
	S_ATTR_FRAMEREL   SymbolRecordKind = 0x112e
	S_ATTR_REGISTER   SymbolRecordKind = 0x112f
	S_ATTR_REGREL     SymbolRecordKind = 0x1130
	S_ATTR_MANYREG    SymbolRecordKind = 0x1131
	S_SEPCODE         SymbolRecordKind = 0x1132
	S_LOCAL_2005      SymbolRecordKind = 0x1133
	S_DEFRANGE_2005   SymbolRecordKind = 0x1134
	S_DEFRANGE2_2005  SymbolRecordKind = 0x1135
	S_SECTION         SymbolRecordKind = 0x1136
	S_COFFGROUP       SymbolRecordKind = 0x1137
	S_EXPORT          SymbolRecordKind = 0x1138
	S_CALLSITEINFO    SymbolRecordKind = 0x1139
	S_FRAMECOOKIE     SymbolRecordKind = 0x113a
	S_DISCARDED       SymbolRecordKind = 0x113b
	S_COMPILE3        SymbolRecordKind = 0x113c
	S_ENVBLOCK        SymbolRecordKind = 0x113d
	S_LOCAL           SymbolRecordKind = 0x113e
	S_DEFRANGE        SymbolRecordKind = 0x113f
	S_DEFRANGE_SUBFIELD SymbolRecordKind = 0x1140
	S_DEFRANGE_REGISTER SymbolRecordKind = 0x1141
	S_DEFRANGE_FRAMEPOINTER_REL SymbolRecordKind = 0x1142
	S_DEFRANGE_SUBFIELD_REGISTER SymbolRecordKind = 0x1143
	S_DEFRANGE_FRAMEPOINTER_REL_FULL_SCOPE SymbolRecordKind = 0x1144
	S_DEFRANGE_REGISTER_REL SymbolRecordKind = 0x1145
	S_LPROC32_ID      SymbolRecordKind = 0x1146
	S_GPROC32_ID      SymbolRecordKind = 0x1147
	S_LPROCMIPS_ID    SymbolRecordKind = 0x1148
	S_GPROCMIPS_ID    SymbolRecordKind = 0x1149
	S_LPROCIA64_ID    SymbolRecordKind = 0x114a
	S_GPROCIA64_ID    SymbolRecordKind = 0x114b
	S_BUILDINFO       SymbolRecordKind = 0x114c
	S_INLINESITE      SymbolRecordKind = 0x114d
	S_INLINESITE_END  SymbolRecordKind = 0x114e
	S_PROC_ID_END     SymbolRecordKind = 0x114f
	S_DEFRANGE_HLSL   SymbolRecordKind = 0x1150
	S_GDATA_HLSL      SymbolRecordKind = 0x1151
	S_LDATA_HLSL      SymbolRecordKind = 0x1152
	S_FILESTATIC      SymbolRecordKind = 0x1153
	S_ARMSWITCHTABLE  SymbolRecordKind = 0x1159
	S_CALLEES         SymbolRecordKind = 0x115a
	S_CALLERS         SymbolRecordKind = 0x115b
	S_POGODATA        SymbolRecordKind = 0x115c
	S_INLINESITE2     SymbolRecordKind = 0x115d
	S_HEAPALLOCSITE   SymbolRecordKind = 0x115e
	S_MOD_TYPEREF     SymbolRecordKind = 0x115f
	S_REF_MINIPDB     SymbolRecordKind = 0x1160
	S_PDBMAP          SymbolRecordKind = 0x1161
	S_GDATA_HLSL32    SymbolRecordKind = 0x1162
	S_LDATA_HLSL32    SymbolRecordKind = 0x1163
	S_GDATA_HLSL32_EX SymbolRecordKind = 0x1164
	S_LDATA_HLSL32_EX SymbolRecordKind = 0x1165
	S_FASTLINK        SymbolRecordKind = 0x1167
	S_INLINEES        SymbolRecordKind = 0x1168
	S_RECTYPE_MAX     SymbolRecordKind = 0x1169
)

// IsProc returns true if this symbol kind represents a procedure.
func (k SymbolRecordKind) IsProc() bool {
	switch k {
	case S_GPROC32, S_LPROC32, S_GPROC32_ID, S_LPROC32_ID,
		S_GPROCIA64, S_LPROCIA64, S_GPROCIA64_ID, S_LPROCIA64_ID,
		S_GPROCMIPS, S_LPROCMIPS, S_GPROCMIPS_ID, S_LPROCMIPS_ID:
		return true
	}
	return false
}

// IsData returns true if this symbol kind represents data.
func (k SymbolRecordKind) IsData() bool {
	switch k {
	case S_GDATA32, S_LDATA32, S_GTHREAD32, S_LTHREAD32:
		return true
	}
	return false
}

// IsPublic returns true if this symbol kind represents a public symbol.
func (k SymbolRecordKind) IsPublic() bool {
	return k == S_PUB32
}

// ProcFlags describes procedure attributes.
type ProcFlags uint8

func (pf ProcFlags) HasFP() bool              { return (pf & 0x01) != 0 }
func (pf ProcFlags) HasIRET() bool            { return (pf & 0x02) != 0 }
func (pf ProcFlags) HasFRET() bool            { return (pf & 0x04) != 0 }
func (pf ProcFlags) IsNoReturn() bool         { return (pf & 0x08) != 0 }
func (pf ProcFlags) IsUnreachable() bool      { return (pf & 0x10) != 0 }
func (pf ProcFlags) HasCustomCallingConv() bool { return (pf & 0x20) != 0 }
func (pf ProcFlags) IsNoInline() bool         { return (pf & 0x40) != 0 }
func (pf ProcFlags) HasOptimizedDebugInfo() bool { return (pf & 0x80) != 0 }

// PublicSymFlags describes public symbol attributes.
type PublicSymFlags uint32

func (psf PublicSymFlags) IsCode() bool     { return (psf & 0x01) != 0 }
func (psf PublicSymFlags) IsFunction() bool { return (psf & 0x02) != 0 }
func (psf PublicSymFlags) IsManaged() bool  { return (psf & 0x04) != 0 }
func (psf PublicSymFlags) IsMSIL() bool     { return (psf & 0x08) != 0 }

// LocalFlags describes local variable attributes.
type LocalFlags uint16

func (lf LocalFlags) IsParameter() bool          { return (lf & 0x0001) != 0 }
func (lf LocalFlags) IsAddressTaken() bool       { return (lf & 0x0002) != 0 }
func (lf LocalFlags) IsCompilerGenerated() bool  { return (lf & 0x0004) != 0 }
func (lf LocalFlags) IsAggregate() bool          { return (lf & 0x0008) != 0 }
func (lf LocalFlags) IsAggregated() bool         { return (lf & 0x0010) != 0 }
func (lf LocalFlags) IsAliased() bool            { return (lf & 0x0020) != 0 }
func (lf LocalFlags) IsAlias() bool              { return (lf & 0x0040) != 0 }
func (lf LocalFlags) IsReturnValue() bool        { return (lf & 0x0080) != 0 }
func (lf LocalFlags) IsOptimizedOut() bool       { return (lf & 0x0100) != 0 }
func (lf LocalFlags) IsEnregisteredGlobal() bool { return (lf & 0x0200) != 0 }
func (lf LocalFlags) IsEnregisteredStatic() bool { return (lf & 0x0400) != 0 }

// SymbolRecord represents a generic symbol record.
type SymbolRecord struct {
	Kind SymbolRecordKind
	Data []byte
}

// ProcSym represents S_GPROC32, S_LPROC32, and related procedure symbols.
type ProcSym struct {
	PtrParent    uint32
	PtrEnd       uint32
	PtrNext      uint32
	CodeSize     uint32
	DbgStart     uint32
	DbgEnd       uint32
	FunctionType tpi.TypeIndex
	CodeOffset   uint32
	Segment      uint16
	Flags        ProcFlags
	Name         string
}

// DataSym represents S_GDATA32, S_LDATA32, S_GTHREAD32, S_LTHREAD32.
type DataSym struct {
	Type    tpi.TypeIndex
	Offset  uint32
	Segment uint16
	Name    string
}

// PublicSym32 represents S_PUB32 (public symbol).
type PublicSym32 struct {
	Flags   PublicSymFlags
	Offset  uint32
	Segment uint16
	Name    string
}

// LocalSym represents S_LOCAL (local variable).
type LocalSym struct {
	Type  tpi.TypeIndex
	Flags LocalFlags
	Name  string
}

// UDTSym represents S_UDT (user-defined type reference).
type UDTSym struct {
	Type tpi.TypeIndex
	Name string
}

// ConstantSym represents S_CONSTANT.
type ConstantSym struct {
	Type  tpi.TypeIndex
	Value uint64
	Name  string
}

// LabelSym represents S_LABEL32.
type LabelSym struct {
	Offset  uint32
	Segment uint16
	Flags   uint8
	Name    string
}

// BlockSym represents S_BLOCK32.
type BlockSym struct {
	PtrParent uint32
	PtrEnd    uint32
	CodeSize  uint32
	Offset    uint32
	Segment   uint16
	Name      string
}

// ThunkSym represents S_THUNK32.
type ThunkSym struct {
	PtrParent uint32
	PtrEnd    uint32
	PtrNext   uint32
	Offset    uint32
	Segment   uint16
	Length    uint16
	Ordinal   uint8
	Name      string
}

// ObjNameSym represents S_OBJNAME.
type ObjNameSym struct {
	Signature uint32
	Name      string
}

// CompileSym3 represents S_COMPILE3.
type CompileSym3 struct {
	Flags       uint32
	Machine     uint16
	FrontendMajor uint16
	FrontendMinor uint16
	FrontendBuild uint16
	FrontendQFE   uint16
	BackendMajor  uint16
	BackendMinor  uint16
	BackendBuild  uint16
	BackendQFE    uint16
	Version     string
}

// RegRelSym represents S_REGREL32.
type RegRelSym struct {
	Offset   uint32
	Type     tpi.TypeIndex
	Register uint16
	Name     string
}

// BPRelSym represents S_BPREL32.
type BPRelSym struct {
	Offset int32
	Type   tpi.TypeIndex
	Name   string
}

// FrameProcSym represents S_FRAMEPROC.
type FrameProcSym struct {
	TotalFrameBytes  uint32
	PaddingFrameBytes uint32
	OffsetToPadding  uint32
	CalleeSaveBytes  uint32
	OffsetOfExceptionHandler int32
	SectionIdOfExceptionHandler uint16
	Flags            uint32
}

// SectionSym represents S_SECTION.
type SectionSym struct {
	SectionNumber uint16
	Alignment     uint8
	Reserved      uint8
	RVA           uint32
	Length        uint32
	Characteristics uint32
	Name          string
}

// CoffGroupSym represents S_COFFGROUP.
type CoffGroupSym struct {
	Size            uint32
	Characteristics uint32
	Offset          uint32
	Segment         uint16
	Name            string
}

// ExportSym represents S_EXPORT.
type ExportSym struct {
	Ordinal uint16
	Flags   uint16
	Name    string
}

// CallSiteInfoSym represents S_CALLSITEINFO.
type CallSiteInfoSym struct {
	Offset     uint32
	Section    uint16
	Padding    uint16
	TypeIndex  tpi.TypeIndex
}

// HeapAllocSiteSym represents S_HEAPALLOCSITE.
type HeapAllocSiteSym struct {
	Offset          uint32
	Section         uint16
	InstructionLen  uint16
	TypeIndex       tpi.TypeIndex
}

// InlineSiteSym represents S_INLINESITE.
type InlineSiteSym struct {
	PtrParent   uint32
	PtrEnd      uint32
	Inlinee     tpi.TypeIndex
	BinaryAnnotations []byte
}

// BuildInfoSym represents S_BUILDINFO.
type BuildInfoSym struct {
	BuildId tpi.TypeIndex
}

// EnvBlockSym represents S_ENVBLOCK.
type EnvBlockSym struct {
	Flags   uint8
	Strings []string
}

// RefSym represents S_PROCREF, S_LPROCREF, S_DATAREF.
type RefSym struct {
	SumName  uint32
	IBSym    uint32
	Imod     uint16
	Name     string
}

// TrampolineSym represents S_TRAMPOLINE.
type TrampolineSym struct {
	Type            uint16
	Size            uint16
	ThunkOffset     uint32
	TargetOffset    uint32
	ThunkSection    uint16
	TargetSection   uint16
}
