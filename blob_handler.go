package goblin

import (
	"fmt"
	"io"
)

type BlobHandler struct {
	Name             string
	Compression      BlockCompression
	CompressionLevel int
	Validate         func(bs []byte) error
	Dump             func(w io.Writer, bs []byte, opts *DumpOpts) error
}

func (h *BlobHandler) GoblinName() string { return h.Name }

func (h *BlobHandler) GoblinDump(w io.Writer, b any, opts *DumpOpts) error {
	blob := b.([]byte)

	if h.Dump != nil {
		return h.Dump(w, blob, opts)
	}

	switch opts.Verbose {
	case DumpSummary:
		fmt.Fprintf(w, "%d bytes\n", len(blob))
	case DumpPreview:
		toWrite := min(64, len(blob))
		h.renderBytes(w, blob[0:toWrite])
		if toWrite < len(blob) {
			fmt.Fprintf(w, "+%d more\n", len(blob)-toWrite)
		}
	case DumpFull:
		h.renderBytes(w, blob)
	default:
		fmt.Fprintf(w, "(unknown verbosity)")
	}

	return nil
}

func (h *BlobHandler) renderBytes(w io.Writer, bs []byte) {
	fmt.Fprintf(w, "%+v\n", bs)
}

func (h *BlobHandler) GoblinValidate(c any) error {
	bs, ok := c.([]byte)
	if !ok {
		return ErrInvalidDataType
	}

	if h.Validate != nil {
		return h.Validate(bs)
	}

	return nil
}

func (h *BlobHandler) GoblinCompression() (BlockCompression, int) {
	return h.Compression, h.CompressionLevel
}

func (h *BlobHandler) GoblinEncode(ec *EncodeContext, w io.Writer, c any) (BlockVersion, error) {
	b, ok := c.([]byte)
	if !ok {
		return 0, ErrInvalidDataType
	}
	_, err := w.Write(b)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (h *BlobHandler) GoblinDecode(dc *DecodeContext, r io.Reader, version BlockVersion, size int64) (any, error) {
	out := make([]byte, size)
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, err
	}
	return out, nil
}
