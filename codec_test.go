package goblin

import (
	"encoding/binary"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testHandler(
	enc func(*EncodeContext, io.Writer, any) (BlockVersion, error),
	dec func(*DecodeContext, io.Reader, BlockVersion, int) (any, error),
) BlockTypeHandler {
	return &th{enc: enc, dec: dec}
}

type th struct {
	enc func(*EncodeContext, io.Writer, any) (BlockVersion, error)
	dec func(*DecodeContext, io.Reader, BlockVersion, int) (any, error)
}

func (h *th) GoblinDump(w io.Writer, b any, opts *DumpOpts) error { return nil }
func (h *th) GoblinLint(c any) error                              { return nil }
func (h *th) GoblinCompression(v BlockVersion) BlockCompression   { return NoCompression }

func (h *th) GoblinEncode(dst *EncodeContext, w io.Writer, c any) (BlockVersion, error) {
	return h.enc(dst, w, c)
}

func (h *th) GoblinDecode(src *DecodeContext, r io.Reader, ver BlockVersion, size int) (any, error) {
	return h.dec(src, r, ver, size)
}

func TestEncodeDecode(t *testing.T) {
	r := NewRegistry()

	type A struct {
		Name         string
		FavouritePet string
		Age          int
	}

	r.RegisterBlockType(0x8000_0001, testHandler(
		func(ec *EncodeContext, w io.Writer, a any) (BlockVersion, error) {
			name, _ := ec.Strings.Add(a.(*A).Name)
			fp, _ := ec.Strings.Add(a.(*A).FavouritePet)
			binary.Write(w, binary.BigEndian, uint32(name))
			binary.Write(w, binary.BigEndian, uint32(fp))
			binary.Write(w, binary.BigEndian, uint32(a.(*A).Age))
			return 1, nil
		},
		func(dc *DecodeContext, r io.Reader, bv BlockVersion, i int) (any, error) {
			var (
				nameRef uint32
				fpRef   uint32
				age     uint32
			)
			binary.Read(r, binary.BigEndian, &nameRef)
			binary.Read(r, binary.BigEndian, &fpRef)
			binary.Read(r, binary.BigEndian, &age)

			name, _ := dc.Strings.Lookup(StringRef(nameRef))
			fp, _ := dc.Strings.Lookup(StringRef(fpRef))

			return &A{
				Name:         name,
				FavouritePet: fp,
				Age:          int(age),
			}, nil
		},
	))

	r.RegisterBlockType(0x8000_5000, testHandler(
		func(ec *EncodeContext, w io.Writer, a any) (BlockVersion, error) {
			if _, err := w.Write(a.([]byte)); err != nil {
				return 0, err
			}
			return 1, nil
		},
		func(dc *DecodeContext, r io.Reader, bv BlockVersion, i int) (any, error) {
			chunk := make([]byte, i)
			if _, err := io.ReadFull(r, chunk); err != nil {
				return nil, err
			}
			return chunk, nil
		},
	))

	cIn := NewContainer()

	cIn.SetBlock(200, 0x8000_0001, "blockA", &A{
		Name:         "Dillon",
		FavouritePet: "Ralph",
		Age:          42,
	})

	cIn.SetBlock(300, 0x8000_5000, "blockB", []byte{1, 2, 3, 4, 5, 6, 7, 8})

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

	dA, ok := cOut.BlockData(200, 0x8000_0001)
	assert.True(t, ok)
	assert.Equal(t, "Dillon", dA.(*A).Name)
	assert.Equal(t, "Ralph", dA.(*A).FavouritePet)
	assert.Equal(t, 42, dA.(*A).Age)

	dB, ok := cOut.BlockData(300, 0x8000_5000)
	assert.True(t, ok)
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8}, dB)
}
