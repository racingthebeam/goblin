package goblin

type FileInfo struct{}

type fileInfoHandler struct{}

func (h *fileInfoHandler) GoblinLint(c any) error {
	return nil
}

func (h *fileInfoHandler) GoblinEncode(ec *EncodeContext, c any) (int, error) {
	return 0, nil
}

func (h *fileInfoHandler) GoblinDecode(dc *DecodeContext, bl int) (any, error) {
	return &FileInfo{}, nil
}
