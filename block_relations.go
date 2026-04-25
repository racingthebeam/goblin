package goblin

import (
	"encoding/binary"
	"fmt"
	"io"
)

type RelationKind uint32

func (rk RelationKind) String() string {
	switch rk {
	case 1:
		return "CONTAINS"
	case 2:
		return "REFERENCES"
	default:
		return "???"
	}
}

const (
	Contains        = RelationKind(1)
	References      = RelationKind(2)
	maxRelationKind = RelationKind(3)

	relationRecordLength = 16
)

type Relation struct {
	FromBlockID BlockID
	ToBlockID   BlockID
	Kind        RelationKind
	Name        string
}

type Relations []Relation

func (r Relations) ChildrenOf(dst []Relation, bid BlockID) []Relation {
	for i := range r {
		if r[i].FromBlockID == bid && r[i].Kind == Contains {
			dst = append(dst, r[i])
		}
	}
	return dst
}

//
// Codec

type relationsHandler struct{}

func (h *relationsHandler) GoblinName() string { return "RELATIONS" }

func (h *relationsHandler) GoblinDump(w io.Writer, b any, opts *DumpOpts) error {
	return nil
}

func (h *relationsHandler) GoblinLint(c any) error {
	rs, ok := c.(Relations)
	if !ok {
		return ErrInvalidDataType
	}

	for i := range rs {
		if rs[i].Kind == 0 || rs[i].Kind >= maxRelationKind {
			return fmt.Errorf("relation %d - invalid kind %d", i, rs[i].Kind)
		} else if len(rs[i].Name) == 0 {
			return fmt.Errorf("relation %d - no name", i)
		}
	}

	return nil
}

func (h *relationsHandler) GoblinCompression(version BlockVersion) BlockCompression {
	return NoCompression
}

func (h *relationsHandler) GoblinEncode(ec *EncodeContext, w io.Writer, c any) (BlockVersion, error) {
	rs, ok := c.(Relations)
	if !ok {
		return 0, ErrInvalidDataType
	}

	for i := range rs {
		str, _ := ec.Strings.Add(rs[i].Name)
		err1 := binary.Write(w, binary.BigEndian, rs[i].FromBlockID)
		err2 := binary.Write(w, binary.BigEndian, rs[i].ToBlockID)
		err3 := binary.Write(w, binary.BigEndian, rs[i].Kind)
		err4 := binary.Write(w, binary.BigEndian, str)
		if err := anyErr(err1, err2, err3, err4); err != nil {
			return 0, err
		}
	}

	return 1, nil
}

func (h *relationsHandler) GoblinDecode(dc *DecodeContext, r io.Reader, version BlockVersion, size int64) (any, error) {
	if size%relationRecordLength != 0 {
		return nil, fmt.Errorf("relation block size must be multiple of %d", relationRecordLength)
	}

	out := make(Relations, size/relationRecordLength)

	for i := range out {
		err1 := binary.Read(r, binary.BigEndian, &out[i].FromBlockID)
		err2 := binary.Read(r, binary.BigEndian, &out[i].ToBlockID)
		err3 := binary.Read(r, binary.BigEndian, &out[i].Kind)
		var str StringRef
		err4 := binary.Read(r, binary.BigEndian, &str)
		if err := anyErr(err1, err2, err3, err4); err != nil {
			return nil, err
		}
		name, found := dc.Strings.Lookup(str)
		if !found {
			return nil, fmt.Errorf("relation %d - invalid string index %d", i, str)
		}
		out[i].Name = name
	}

	return out, nil
}
