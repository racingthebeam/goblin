package goblin

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Strings struct {
	strings []string
	lookup  map[string]StringRef
}

func NewStrings() *Strings {
	return &Strings{
		strings: make([]string, 0),
		lookup:  map[string]StringRef{},
	}
}

func (s *Strings) All() []string {
	return s.strings
}

func (s *Strings) Has(str string) bool {
	if len(str) == 0 {
		return true
	}

	_, ok := s.lookup[str]
	return ok
}

func (s *Strings) Add(str string) (StringRef, bool) {
	if len(str) == 0 {
		return 0, true
	}

	ix, exists := s.lookup[str]
	if exists {
		return ix, true
	}

	ix = StringRef(len(s.strings) + 1)
	s.lookup[str] = ix
	s.strings = append(s.strings, str)

	return ix, false
}

func (s *Strings) Lookup(i StringRef) (string, bool) {
	if i == 0 {
		return "", true
	}

	if i > StringRef(len(s.strings)) {
		return "", false
	}

	return s.strings[i-1], true
}

type stringsHandler struct{}

func (h *stringsHandler) GoblinName() string { return "STRINGS" }

func (h *stringsHandler) GoblinDump(w io.Writer, b any, opts *DumpOpts) error {
	return nil
}

func (h *stringsHandler) GoblinLint(c any) error {
	ss, ok := c.(*Strings)
	if !ok {
		return ErrInvalidDataType
	}
	for ix, str := range ss.strings {
		if nix := strings.IndexByte(str, 0); nix >= 0 {
			return fmt.Errorf("string at %d contains illegal null byte at index %d", ix, nix)
		}
	}
	return nil
}

func (h *stringsHandler) GoblinCompression(version BlockVersion) BlockCompression {
	return NoCompression
}

var nul = []byte{0}

func (h *stringsHandler) GoblinEncode(ec *EncodeContext, w io.Writer, c any) (BlockVersion, error) {
	ss, ok := c.(*Strings)
	if !ok {
		return 0, ErrInvalidDataType
	}

	for ix, str := range ss.strings {
		if _, err := w.Write([]byte(str)); err != nil {
			return 0, fmt.Errorf("failed to write string at index %d (%s)", ix, err)
		} else if _, err := w.Write(nul); err != nil {
			return 0, fmt.Errorf("failed to write null terminator at index %d (%s)", ix, err)
		}
	}

	return 1, nil
}

func (h *stringsHandler) GoblinDecode(dc *DecodeContext, r io.Reader, version BlockVersion, size int) (any, error) {
	r = &io.LimitedReader{R: r, N: int64(size)}

	s := bufio.NewScanner(r)
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		} else if i := bytes.IndexByte(data, 0); i >= 0 {
			return i + 1, data[0:i], nil
		} else if atEOF {
			return 0, nil, errors.New("unterminated string at end of input")
		}
		return 0, nil, nil
	})

	out := NewStrings()
	for s.Scan() {
		_, exists := out.Add(s.Text())
		if exists {
			return nil, fmt.Errorf("duplicate string")
		}
	}

	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("string scan failed (%s)", err)
	}

	return out, nil
}
