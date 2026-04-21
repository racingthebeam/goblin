package goblin

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
)

func anyErr(errors ...error) error {
	for _, e := range errors {
		if e != nil {
			return e
		}
	}
	return nil
}

func wrapWriter(w io.Writer, c BlockCompression) (io.WriteCloser, error) {
	switch c {
	case NoCompression:
		return &nullWriteCloser{w: w}, nil
	case GZip:
		return gzip.NewWriter(w), nil
	case ZLib:
		return zlib.NewWriter(w), nil
	default:
		return nil, fmt.Errorf("unknown compression type %d", c)
	}
}

type nullWriteCloser struct {
	w io.Writer
}

func (w *nullWriteCloser) Write(p []byte) (int, error) { return w.w.Write(p) }
func (w *nullWriteCloser) Close() error                { return nil }
