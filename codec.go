package goblin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
)

var (
	header = []byte{
		0x7F, // non-printable
		0x47, // G
		0x4F, // O
		0x42, // B
		0x4C, // L
		0x49, // I
		0x4E, // N
		0x01, // version number (1)
	}

	indexZeroes = [indexEntrySize]byte{}
)

type indexEntry struct {
	ID          BlockID          // 4
	Type        BlockType        // 4
	Name        StringRef        // 4
	Version     BlockVersion     // 2
	Compression BlockCompression // 2
	Offset      int64            // 8
	Size        int64            // 8
}

const (
	indexEntrySize = 32
)

type Option func(any)

func WithRegistry(r *Registry) Option {
	return func(thing any) {
		if enc, ok := thing.(*Encoder); ok {
			enc.reg = r
		} else if dec, ok := thing.(*Decoder); ok {
			dec.reg = r
		}
	}
}

//
// Encoder

type Encoder struct {
	w io.WriteSeeker

	index []indexEntry
	c     *EncodeContext
	reg   *Registry
}

func NewEncoder(w io.WriteSeeker, opts ...Option) *Encoder {
	e := &Encoder{w: w}
	for _, o := range opts {
		o(e)
	}
	return e
}

func (e *Encoder) Encode(c *Container) error {
	// 2 extra blocks - strings and relations
	blockCount := len(c.blocks) + 2

	e.index = make([]indexEntry, 0, blockCount)
	e.c = NewEncodeContext()

	if _, err := e.w.Write(header); err != nil {
		return err
	}

	binary.Write(e.w, binary.BigEndian, uint32(blockCount))

	indexOffset, err := e.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	// can't just seek forward because underlying target might not be a file
	for range blockCount {
		if _, err := e.w.Write(indexZeroes[:]); err != nil {
			return err
		}
	}

	for _, blockID := range slices.Sorted(maps.Keys(c.blocks)) {
		block := c.blocks[blockID]
		if ent, err := e.writeBlock(block); err != nil {
			return err
		} else {
			e.index = append(e.index, ent)
		}
	}

	if ent, err := e.writeBlock(&Block{
		ID:   BlockIDRelations,
		Type: BlockTypeRelations,
		Data: c.relations,
	}); err != nil {
		return fmt.Errorf("failed to write relations block (%s)", err)
	} else {
		e.index = append(e.index, ent)
	}

	if ent, err := e.writeBlock(&Block{
		ID:   BlockIDStrings,
		Type: BlockTypeStrings,
		Data: e.c.Strings,
	}); err != nil {
		return fmt.Errorf("failed to write strings block (%s)", err)
	} else {
		e.index = append(e.index, ent)
	}

	if err := e.writeIndex(indexOffset); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) writeBlock(b *Block) (indexEntry, error) {
	hnd, found := e.reg.LookupBlockType(b.Type)
	if !found {
		return indexEntry{}, fmt.Errorf("no block type handler found for block type %d", b.Type)
	}

	ent := indexEntry{ID: b.ID, Type: b.Type}

	name, _ := e.c.Strings.Add(b.Name)
	ent.Name = name

	offset, err := e.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return ent, err
	}

	ent.Offset = offset

	comp := hnd.GoblinCompression(0)
	w, err := wrapWriter(e.w, comp)
	if err != nil {
		return indexEntry{}, err
	}

	ent.Compression = comp

	if version, err := hnd.GoblinEncode(e.c, w, b.Data); err != nil {
		return indexEntry{}, err
	} else if err := w.Close(); err != nil {
		return indexEntry{}, err
	} else {
		ent.Version = version
	}

	offset, err = e.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return indexEntry{}, err
	}

	ent.Size = offset - ent.Offset

	if err := e.align4(); err != nil {
		return indexEntry{}, err
	}

	return ent, nil
}

func (e *Encoder) writeIndex(offset int64) error {
	if _, err := e.w.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	for i, ie := range e.index {
		err1 := binary.Write(e.w, binary.BigEndian, ie.ID)
		err2 := binary.Write(e.w, binary.BigEndian, ie.Type)
		err3 := binary.Write(e.w, binary.BigEndian, ie.Name)
		err4 := binary.Write(e.w, binary.BigEndian, ie.Version)
		err5 := binary.Write(e.w, binary.BigEndian, ie.Compression)
		err6 := binary.Write(e.w, binary.BigEndian, ie.Offset)
		err7 := binary.Write(e.w, binary.BigEndian, ie.Size)
		if err := anyErr(err1, err2, err3, err4, err5, err6, err7); err != nil {
			return fmt.Errorf("failed to write index entry %d (%s)", i, err)
		}
	}

	return nil
}

func (e *Encoder) align4() error {
	off, err := e.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	n := 4 - (off % 4)
	if n < 4 {
		if _, err := e.w.Write(indexZeroes[:n]); err != nil {
			return err
		}
	}
	return nil
}

//
// EncodeContext

type EncodeContext struct {
	Strings *Strings
}

func NewEncodeContext() *EncodeContext {
	return &EncodeContext{
		Strings: NewStrings(),
	}
}

//
// Decoder

type Decoder struct {
	r   io.ReadSeeker
	c   *DecodeContext
	reg *Registry
}

func NewDecoder(r io.ReadSeeker, opts ...Option) *Decoder {
	d := &Decoder{r: r}
	for _, o := range opts {
		o(d)
	}
	return d
}

func (d *Decoder) Decode() (*Container, error) {
	out := NewContainer()

	buf := make([]byte, indexEntrySize)

	n, err := d.r.Read(buf[0:12])
	if err != nil {
		return nil, fmt.Errorf("failed to read container header (%s)", err)
	} else if n != 12 {
		return nil, fmt.Errorf("expected 12 bytes, got %d", n)
	} else if !bytes.Equal(header, buf[0:8]) {
		return nil, errors.New("invalid header")
	}

	index := make([]indexEntry, binary.BigEndian.Uint32(buf[8:]))

	for i := range len(index) {
		if err := d.readIndexEntry(&index[i], buf); err != nil {
			return nil, fmt.Errorf("failed to read index entry %d (%s)", i, err)
		}
	}

	found := false
	for i := range index {
		if index[i].Type != BlockTypeStrings {
			continue
		}
		block, err := d.readBlockFromEntry(&index[i])
		if err != nil {
			return nil, fmt.Errorf("failed to read strings block (%s)", err)
		}
		asStrings, ok := block.Data.(*Strings)
		if !ok {
			return nil, errors.New("strings block returned incorrect data type")
		}
		d.c = NewDecodeContext(asStrings)
		found = true
	}

	if !found {
		return nil, errors.New("strings block not found")
	}

	for i := range index {
		if index[i].Type == BlockTypeStrings {
			continue
		} else if block, err := d.readBlockFromEntry(&index[i]); err != nil {
			return nil, err
		} else if block.Type == BlockTypeRelations {
			if rs, ok := block.Data.(Relations); !ok {
				return nil, errors.New("relations block returned incorrect data type")
			} else {
				out.relations = rs
			}
		} else {
			out.blocks[block.ID] = block
		}
	}

	return out, nil
}

func (d *Decoder) readIndexEntry(dst *indexEntry, buf []byte) error {
	if _, err := io.ReadFull(d.r, buf); err != nil {
		return fmt.Errorf("failed to read index entry (%s)", err)
	}
	dst.ID = BlockID(binary.BigEndian.Uint32(buf[0:4]))
	dst.Type = BlockType(binary.BigEndian.Uint32(buf[4:8]))
	dst.Name = StringRef(binary.BigEndian.Uint32(buf[8:12]))
	dst.Version = BlockVersion(binary.BigEndian.Uint16(buf[12:14]))
	dst.Compression = BlockCompression(binary.BigEndian.Uint16(buf[14:16]))
	dst.Offset = int64(binary.BigEndian.Uint64(buf[16:24]))
	dst.Size = int64(binary.BigEndian.Uint64(buf[24:32]))
	return nil
}

func (d *Decoder) readBlockFromEntry(ent *indexEntry) (*Block, error) {
	hnd, found := d.reg.LookupBlockType(ent.Type)
	if !found {
		return nil, fmt.Errorf("block ID %d has unknown type %d", ent.ID, ent.Type)
	}

	_, err := d.r.Seek(ent.Offset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to start of block ID %d (%s)", ent.ID, err)
	}

	r, err := wrapReader(&io.LimitedReader{R: d.r, N: ent.Size}, ent.Compression)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap reader for block ID %d (%s)", ent.ID, err)
	}

	data, err := hnd.GoblinDecode(d.c, r, ent.Version, int(ent.Size))
	if err != nil {
		return nil, fmt.Errorf("block ID %d decode failed (%s)", ent.ID, err)
	}

	// If the strings table is not populated it means we're reading the strings
	// block itself. Just skip the lookup since it never has a name.
	name := ""
	if d.c != nil {
		name, found = d.c.Strings.Lookup(ent.Name)
		if !found {
			return nil, fmt.Errorf("block ID %d name not found in strings table", ent.ID)
		}
	}

	return &Block{
		ID:   ent.ID,
		Type: ent.Type,
		Name: name,
		Data: data,
	}, nil
}

//
// DecodeContext

type DecodeContext struct {
	Strings *Strings
}

func NewDecodeContext(strings *Strings) *DecodeContext {
	return &DecodeContext{
		Strings: strings,
	}
}
