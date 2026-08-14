package goblin

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

const (
	mtUint8   = 1
	mtUint16  = 2
	mtUint32  = 3
	mtUint64  = 4
	mtInt8    = 5
	mtInt16   = 6
	mtInt32   = 7
	mtInt64   = 8
	mtFloat32 = 9
	mtFloat64 = 10
	mtBool    = 11
	mtString  = 12
)

var mdTypeReflection = map[byte]reflect.Type{
	mtUint8:   reflect.TypeFor[uint8](),
	mtUint16:  reflect.TypeFor[uint16](),
	mtUint32:  reflect.TypeFor[uint32](),
	mtUint64:  reflect.TypeFor[uint64](),
	mtInt8:    reflect.TypeFor[int8](),
	mtInt16:   reflect.TypeFor[int16](),
	mtInt32:   reflect.TypeFor[int32](),
	mtInt64:   reflect.TypeFor[int64](),
	mtFloat32: reflect.TypeFor[float32](),
	mtFloat64: reflect.TypeFor[float64](),
	mtBool:    reflect.TypeFor[bool](),
}

var ErrMetadataTooLong = errors.New("metadata too long")
var ErrUnsupportedMetadataValue = errors.New("unsupported metadata value")

func metadataValidateValue(v any) error {
	switch v.(type) {
	case uint8, uint16, uint32, uint64, int8, int16, int32, int64, float32, float64, bool, string:
		return nil
	default:
		return ErrUnsupportedMetadataValue
	}
}

func metadataDecode(src io.Reader, st *Strings) (map[string]any, error) {
	count := uint16(0)
	if err := read(src, &count); err != nil {
		return nil, err
	}

	out := make(map[string]any, count)

	for range count {
		keyIx := StringRef(0)
		if err := read(src, &keyIx); err != nil {
			return nil, err
		}

		key, ok := st.Lookup(keyIx)
		if !ok {
			return nil, fmt.Errorf("invalid string table entry %d", keyIx)
		}

		tag := byte(0)
		if err := read(src, &tag); err != nil {
			return nil, err
		}

		if tag == mtString {
			var ref StringRef
			if err := read(src, &ref); err != nil {
				return nil, err
			}
			str, ok := st.Lookup(ref)
			if !ok {
				return nil, fmt.Errorf("invalid string table entry %d", ref)
			}
			out[key] = str
		} else {
			rt, ok := mdTypeReflection[tag]
			if !ok {
				return nil, fmt.Errorf("unknown metadata type tag %d", tag)
			}
			ptr := reflect.New(rt)
			if err := read(src, ptr.Interface()); err != nil {
				return nil, err
			}
			out[key] = ptr.Elem().Interface()
		}
	}

	return out, nil
}

func metadataEncode(dst io.Writer, st *Strings, metadata map[string]any) error {
	if len(metadata) > 65535 {
		return ErrMetadataTooLong
	}

	for k, v := range metadata {
		if err := metadataValidateValue(v); err != nil {
			return fmt.Errorf("%w for key %q", err, k)
		}
	}

	if err := write(dst, uint16(len(metadata))); err != nil {
		return err
	}

	for k, v := range metadata {
		tag := byte(0)
		switch vActual := v.(type) {
		case uint8:
			tag = mtUint8
		case uint16:
			tag = mtUint16
		case uint32:
			tag = mtUint32
		case uint64:
			tag = mtUint64
		case int8:
			tag = mtInt8
		case int16:
			tag = mtInt16
		case int32:
			tag = mtInt32
		case int64:
			tag = mtInt64
		case float32:
			tag = mtFloat32
		case float64:
			tag = mtFloat64
		case bool:
			tag = mtBool
		case string:
			tag = mtString
			v, _ = st.Add(vActual)
		}

		strIx, _ := st.Add(k)
		if err := write(dst, strIx); err != nil {
			return err
		} else if err := write(dst, tag); err != nil {
			return err
		} else if err := write(dst, v); err != nil {
			return err
		}
	}

	return nil
}
