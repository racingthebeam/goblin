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

func wrapReader(r io.Reader, c BlockCompression) (io.ReadCloser, error) {
	switch c {
	case NoCompression:
		return &nullReadCloser{r: r}, nil
	case GZip:
		return gzip.NewReader(r)
	case ZLib:
		return zlib.NewReader(r)
	default:
		return nil, fmt.Errorf("unknown compression type %d", c)
	}
}

type nullReadCloser struct {
	r io.Reader
}

func (r *nullReadCloser) Read(p []byte) (int, error) { return r.r.Read(p) }
func (r *nullReadCloser) Close() error               { return nil }
