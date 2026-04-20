package goblin

import (
	"errors"
)

var (
	ErrInvalidDataType = errors.New("invalid data type")
)

type BlockID uint32

type BlockType uint32

func (bt BlockType) IsPublic() bool  { return bt&BlockType(0x8000_0000) == 0 }
func (bt BlockType) IsPrivate() bool { return !bt.IsPublic() }

// Built-in block types
const (
	BlockTypeFileInfo  = BlockType(1)
	BlockTypeMetadata  = BlockType(2)
	BlockTypeRelations = BlockType(3)
	BlockTypeStrings   = BlockType(4)
	BlockTypeBlob      = BlockType(5)
)

type StringRef uint32

type BlockContent interface {
	// Returns the type of this block
	GoblinType() BlockType
}

type BlockTypeHandler interface {
	// TODO: GoblinDump(w io.Writer, opts *DumpOpts)
	GoblinLint(c any) error
	GoblinEncode(dst *EncodeContext, c any) (int, error)
	GoblinDecode(src *DecodeContext, size int) (any, error)
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
