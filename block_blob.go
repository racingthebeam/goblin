package goblin

import "io"

type Blob []byte

type blobHandler struct{}

func (h *blobHandler) GoblinLint(c any) error {
	_, ok := c.(Blob)
	if !ok {
		return ErrInvalidDataType
	}
	return nil
}

func (h *blobHandler) GoblinEncode(ec *EncodeContext, c any) (int, error) {
	b, ok := c.(Blob)
	if !ok {
		return 0, ErrInvalidDataType
	}
	return ec.Write(b)
}

func (h *blobHandler) GoblinDecode(dc *DecodeContext, n int) (any, error) {
	out := make(Blob, n)
	if _, err := io.ReadFull(dc, out); err != nil {
		return nil, err
	}
	return out, nil
}
