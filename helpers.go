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

func wrapWriter(w io.Writer, c BlockCompression) (*countingWriter, error) {
	switch c {
	case NoCompression:
		return &countingWriter{w: w}, nil
	case GZip:
		return &countingWriter{w: gzip.NewWriter(w), c: true}, nil
	case ZLib:
		return &countingWriter{w: zlib.NewWriter(w), c: true}, nil
	default:
		return nil, fmt.Errorf("unknown compression type %d", c)
	}
}

type countingWriter struct {
	Written int
	w       io.Writer
	c       bool
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.Written += n
	return n, err
}

func (w *countingWriter) Close() error {
	if wc, ok := w.w.(io.WriteCloser); w.c && ok {
		return wc.Close()
	}
	return nil
}

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
