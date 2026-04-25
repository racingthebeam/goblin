package goblin

import "io"

type FileInfo struct{}

type fileInfoHandler struct{}

func (h *fileInfoHandler) GoblinName() string { return "FILEINFO" }

func (h *fileInfoHandler) GoblinDump(w io.Writer, b any, opts *DumpOpts) error {
	return nil
}

func (h *fileInfoHandler) GoblinLint(c any) error {
	return nil
}

func (h *fileInfoHandler) GoblinCompression(version BlockVersion) BlockCompression {
	return NoCompression
}

func (h *fileInfoHandler) GoblinEncode(ec *EncodeContext, w io.Writer, c any) (BlockVersion, error) {
	return 1, nil
}

func (h *fileInfoHandler) GoblinDecode(dc *DecodeContext, r io.Reader, version BlockVersion, size int64) (any, error) {
	return &FileInfo{}, nil
}
