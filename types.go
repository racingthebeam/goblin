package goblin

import (
	"errors"
	"io"
)

var (
	ErrInvalidDataType = errors.New("invalid data type")
)

type BlockID uint32

func (i BlockID) IsReserved() bool { return i >= 0xFFFF_FF00 }

const (
	BlockIDRelations = BlockID(0xFFFF_FF00)
	BlockIDStrings   = BlockID(0xFFFF_FFFF)
)

type BlockType uint32

// Built-in block types
const (
	BlockTypeFileInfo  = BlockType(1)
	BlockTypeStrings   = BlockType(2)
	BlockTypeRelations = BlockType(3)
	BlockTypeMetadata  = BlockType(4)
	BlockTypeBlob      = BlockType(5)
)

func (bt BlockType) IsPublic() bool  { return bt&BlockType(0x8000_0000) == 0 }
func (bt BlockType) IsPrivate() bool { return !bt.IsPublic() }

type BlockVersion uint16

type BlockCompression uint16

func (c BlockCompression) String() string {
	switch c {
	case NoCompression:
		return "none"
	case GZip:
		return "gzip"
	case ZLib:
		return "zlib"
	default:
		return "????"
	}
}

const (
	NoCompression = BlockCompression(0)
	GZip          = BlockCompression(1)
	ZLib          = BlockCompression(2)
)

type StringRef uint32

type BlockContent interface {
	// Returns the type of this block
	GoblinType() BlockType
}

type IndexEntry struct {
	ID             BlockID          // 4
	Type           BlockType        // 4
	Name           StringRef        // 4
	Version        BlockVersion     // 2
	Compression    BlockCompression // 2
	Offset         int64            // 8
	DataSize       uint32           // 4
	CompressedSize uint32           // 4
}

type BlockTypeHandler interface {
	GoblinName() string

	GoblinDump(w io.Writer, b any, opts *DumpOpts) error

	GoblinLint(c any) error

	// Returns the compression type employed by the given version number.
	// This method is used to wrap to readers/writers passed into GoblinDecode()
	// and GoblinEncode().
	//
	// When writing, version will be zero, so the method should return
	// whatever compression type is required for newly written blocks
	// i.e. that matching the version number ultimately returned by
	// GoblinEncode().
	GoblinCompression(version BlockVersion) BlockCompression

	// Encode the block to the target writer, returning the version number.
	GoblinEncode(dst *EncodeContext, w io.Writer, c any) (BlockVersion, error)

	GoblinDecode(src *DecodeContext, r io.Reader, ver BlockVersion, size int64) (any, error)
}

const (
	// Dump summary only (e.g. info about size, encoding)
	DumpSummary = 0

	// Dump small preview (e.g. first few lines of text)
	DumpPreview = 1

	// Dump full contents
	DumpFull = 2
)

type DumpOpts struct {
	Color   bool // Colorize output?
	Verbose int  // Verbosity level; see Dump* constants
}
