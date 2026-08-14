package goblin

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadataEncodeDecode(t *testing.T) {
	metadata := map[string]any{
		"uint8":  uint8(1),
		"uint16": uint16(2),
		"uint32": uint32(3),
		"uint64": uint64(4),

		"int8":  int8(-1),
		"int16": int16(-2),
		"int32": int32(-3),
		"int64": int64(-4),

		"float32": float32(1.234),
		"float64": float64(10.12345),

		"bool": true,

		"string": "hello there",
	}

	st := NewStrings()

	buf := bytes.Buffer{}

	err := metadataEncode(&buf, st, metadata)
	assert.NoError(t, err)

	decoded, err := metadataDecode(&buf, st)
	assert.NoError(t, err)

	assert.Equal(t, metadata, decoded)
}
