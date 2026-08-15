package goblin

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testHandler(
	name string,
	md func(any) (map[string]any, error),
	enc func(*EncodeContext, io.Writer, any) (BlockVersion, error),
	dec func(*DecodeContext, io.Reader, BlockVersion, int64, map[string]any) (any, error),
) BlockTypeHandler {
	return &th{name: name, md: md, enc: enc, dec: dec}
}

type th struct {
	name string
	md   func(any) (map[string]any, error)
	enc  func(*EncodeContext, io.Writer, any) (BlockVersion, error)
	dec  func(*DecodeContext, io.Reader, BlockVersion, int64, map[string]any) (any, error)
}

func (h *th) GoblinName() string                                  { return h.name }
func (h *th) GoblinDump(w io.Writer, b any, opts *DumpOpts) error { return nil }
func (h *th) GoblinValidate(c any) error                          { return nil }
func (h *th) GoblinCompression() (BlockCompression, int)          { return NoCompression, 0 }
func (h *th) GoblinMetadata(b any) (map[string]any, error)        { return h.md(b) }

func (h *th) GoblinEncode(dst *EncodeContext, w io.Writer, c any) (BlockVersion, error) {
	return h.enc(dst, w, c)
}

func (h *th) GoblinDecode(src *DecodeContext, r io.Reader, ver BlockVersion, size int64, metadata map[string]any) (any, error) {
	return h.dec(src, r, ver, size, metadata)
}

func TestEncodeDecode(t *testing.T) {
	r := NewRegistry()

	type A struct {
		Name         string
		FavouritePet string
		Age          int
		Nickname     string
	}

	r.RegisterBlockType(0x0000_0001, testHandler(
		"testA",
		func(a any) (map[string]any, error) {
			instance := a.(*A)
			return map[string]any{
				"nickname": instance.Nickname,
			}, nil
		},
		func(ec *EncodeContext, w io.Writer, a any) (BlockVersion, error) {
			instance := a.(*A)
			ec.WriteRecord(w, instance.Name, instance.FavouritePet, uint32(instance.Age))
			return 1, nil
		},
		func(dc *DecodeContext, r io.Reader, bv BlockVersion, i int64, md map[string]any) (any, error) {
			out := A{}
			var age uint32
			if err := dc.ReadRecord(r, &out.Name, &out.FavouritePet, &age); err != nil {
				return nil, err
			}
			out.Age = int(age)
			out.Nickname = md["nickname"].(string)
			return &out, nil
		},
	))

	r.RegisterBlockType(0x0000_5000, testHandler(
		"testB",
		func(a any) (map[string]any, error) {
			return nil, nil
		},
		func(ec *EncodeContext, w io.Writer, a any) (BlockVersion, error) {
			if _, err := w.Write(a.([]byte)); err != nil {
				return 0, err
			}
			return 1, nil
		},
		func(dc *DecodeContext, r io.Reader, bv BlockVersion, i int64, md map[string]any) (any, error) {
			chunk := make([]byte, i)
			if _, err := io.ReadFull(r, chunk); err != nil {
				return nil, err
			}
			return chunk, nil
		},
	))

	cIn := NewContainer(WithFileType(123))

	cIn.SetBlock(200, 0x0000_0001, "blockA", &A{
		Name:         "Dillon",
		FavouritePet: "Ralph",
		Age:          42,
		Nickname:     "RD",
	})

	cIn.SetBlock(300, 0x0000_5000, "blockB", []byte{1, 2, 3, 4, 5, 6, 7, 8})

	tmp, err := os.CreateTemp("", "goblintest")
	if err != nil {
		t.Fatalf("failed to create temporary file (%s)", err)
	}

	t.Cleanup(func() { os.Remove(tmp.Name()) })

	if err := NewEncoder(tmp, WithRegistry(r)).Encode(cIn); err != nil {
		t.Fatalf("encode failed: %s", err)
	}

	tmp.Seek(0, io.SeekStart)

	cOut, err := NewDecoder(tmp, WithRegistry(r)).Decode()
	if err != nil {
		t.Fatalf("decode failed: %s", err)
	}

	assert.Equal(t, FileType(123), cOut.FileType)

	dA, ok := cOut.BlockData(200, 0x0000_0001)
	assert.True(t, ok)
	assert.Equal(t, "Dillon", dA.(*A).Name)
	assert.Equal(t, "Ralph", dA.(*A).FavouritePet)
	assert.Equal(t, 42, dA.(*A).Age)
	assert.Equal(t, "RD", dA.(*A).Nickname)

	dB, ok := cOut.BlockData(300, 0x0000_5000)
	assert.True(t, ok)
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8}, dB)
}
