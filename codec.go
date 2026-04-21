package goblin

import (
	"encoding/binary"
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

//
// Encoder

type Encoder struct {
	w io.WriteSeeker

	index []indexEntry
	c     *EncodeContext
}

func NewEncoder(w io.WriteSeeker) *Encoder {
	return &Encoder{w: w}
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
	hnd, found := LookupBlockType(b.Type)
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

//
// DecodeContext

type DecodeContext struct {
	Strings *Strings
}

func NewDecodeContext(r io.Reader, strings *Strings) *DecodeContext {
	return &DecodeContext{
		Strings: strings,
	}
}
