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

func wrapWriter(w io.Writer, c BlockCompression, level int) (*countingWriter, error) {
	switch c {
	case NoCompression:
		return &countingWriter{w: w}, nil
	case GZip:
		cw, err := gzip.NewWriterLevel(w, level)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip writer (%s)", err)
		}
		return &countingWriter{w: cw, close: true}, nil
	case ZLib:
		cw, err := zlib.NewWriterLevel(w, level)
		if err != nil {
			return nil, fmt.Errorf("failed to create zlib writer (%s)", err)
		}
		return &countingWriter{w: cw, close: true}, nil
	default:
		return nil, fmt.Errorf("unknown compression type %d", c)
	}
}

type countingWriter struct {
	Written int
	w       io.Writer
	close   bool
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.Written += n
	return n, err
}

func (w *countingWriter) Close() error {
	if wc, ok := w.w.(io.WriteCloser); w.close && ok {
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
