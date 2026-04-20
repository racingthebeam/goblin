package goblin

import "io"

type EncodeContext struct {
	Strings *Strings
	w       io.Writer
}

func NewEncodeContext(w io.Writer) *EncodeContext {
	return &EncodeContext{
		Strings: NewStrings(),
		w:       w,
	}
}

func (ec *EncodeContext) Write(p []byte) (int, error) {
	return ec.w.Write(p)
}

type DecodeContext struct {
	Strings *Strings

	r io.Reader
}

func NewDecodeContext(r io.Reader, strings *Strings) *DecodeContext {
	return &DecodeContext{
		Strings: strings,
		r:       r,
	}
}

func (dc *DecodeContext) Read(p []byte) (int, error) {
	return dc.r.Read(p)
}
