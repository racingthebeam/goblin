package goblin

import (
	"fmt"
	"io"
)

type Blob []byte

type blobHandler struct{}

func (h *blobHandler) GoblinName() string { return "BLOB" }

func (h *blobHandler) GoblinDump(w io.Writer, b any, opts *DumpOpts) error {
	blob := b.(Blob)

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

func (h *blobHandler) renderBytes(w io.Writer, bs []byte) {
	fmt.Fprintf(w, "%+v\n", bs)
}

func (h *blobHandler) GoblinLint(c any) error {
	_, ok := c.(Blob)
	if !ok {
		return ErrInvalidDataType
	}
	return nil
}

func (h *blobHandler) GoblinCompression(version BlockVersion) BlockCompression {
	return NoCompression
}

func (h *blobHandler) GoblinEncode(ec *EncodeContext, w io.Writer, c any) (BlockVersion, error) {
	b, ok := c.(Blob)
	if !ok {
		return 0, ErrInvalidDataType
	}
	_, err := w.Write(b)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (h *blobHandler) GoblinDecode(dc *DecodeContext, r io.Reader, version BlockVersion, size int64) (any, error) {
	out := make(Blob, size)
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, err
	}
	return out, nil
}
