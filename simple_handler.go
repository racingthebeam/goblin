package goblin

import (
	"fmt"
	"io"
)

type SimpleHandler struct {
	Name        string
	Dump        func(w io.Writer, b any, opts *DumpOpts) error
	Validate    func(b any) error
	Compression func() (BlockCompression, int)
	Encode      func(dst *EncodeContext, w io.Writer, b any) (BlockVersion, error)
	Decode      func(src *DecodeContext, r io.Reader, ver BlockVersion, size int64) (any, error)
	Decoders    map[BlockVersion]func(src *DecodeContext, r io.Reader, ver BlockVersion, size int64) (any, error)
}

func (h *SimpleHandler) GoblinName() string { return h.Name }

func (h *SimpleHandler) GoblinDump(w io.Writer, b any, opts *DumpOpts) error {
	if h.Dump != nil {
		return h.Dump(w, b, opts)
	}
	return nil
}

func (h *SimpleHandler) GoblinValidate(b any) error {
	if h.Validate != nil {
		return h.Validate(b)
	}
	return nil
}

func (h *SimpleHandler) GoblinCompression() (BlockCompression, int) {
	if h.Compression != nil {
		return h.Compression()
	}
	return NoCompression, 0
}

func (h *SimpleHandler) GoblinEncode(dst *EncodeContext, w io.Writer, b any) (BlockVersion, error) {
	return h.Encode(dst, w, b)
}

func (h *SimpleHandler) GoblinDecode(src *DecodeContext, r io.Reader, ver BlockVersion, size int64) (any, error) {
	dec := h.Decode
	if h.Decoders != nil && h.Decoders[ver] != nil {
		dec = h.Decoders[ver]
	}
	if dec == nil {
		return nil, fmt.Errorf("no suitable decoder found for version %d", ver)
	}
	return dec(src, r, ver, size)
}
