package goblin

import "io"

type Blob []byte

type blobHandler struct{}

func (h *blobHandler) GoblinDump(w io.Writer, b any, opts *DumpOpts) error {
	return nil
}

func (h *blobHandler) GoblinLint(c any) error {
	_, ok := c.(Blob)
	if !ok {
		return ErrInvalidDataType
	}
	return nil
}

func (h *blobHandler) GoblinCompression(version BlockVersion) BlockCompression { return NoCompression }

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

func (h *blobHandler) GoblinDecode(dc *DecodeContext, r io.Reader, version BlockVersion, size int) (any, error) {
	out := make(Blob, size)
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, err
	}
	return out, nil
}
