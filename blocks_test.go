package goblin

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringsCodec(t *testing.T) {
	ss := NewStrings()

	var i StringRef
	var ex bool

	i, ex = ss.Add("foo")
	assert.Equal(t, StringRef(1), i)
	assert.False(t, ex)

	i, ex = ss.Add("bar")
	assert.Equal(t, StringRef(2), i)
	assert.False(t, ex)

	i, ex = ss.Add("quux")
	assert.Equal(t, StringRef(3), i)
	assert.False(t, ex)

	i, ex = ss.Add("bar")
	assert.Equal(t, StringRef(2), i)
	assert.True(t, ex)

	var buf bytes.Buffer

	codec := &stringsHandler{}

	ec := NewEncodeContext()

	n, err := codec.GoblinEncode(ec, &buf, ss)
	assert.NoError(t, err)
	assert.Equal(t, BlockVersion(1), n)

	dc := NewDecodeContext(&buf, nil)
	reloaded, err := codec.GoblinDecode(dc, &buf, 1, 13)
	assert.NoError(t, err)

	str, ok := reloaded.(*Strings).Lookup(1)
	assert.True(t, ok)
	assert.Equal(t, "foo", str)

	str, ok = reloaded.(*Strings).Lookup(2)
	assert.True(t, ok)
	assert.Equal(t, "bar", str)

	str, ok = reloaded.(*Strings).Lookup(3)
	assert.True(t, ok)
	assert.Equal(t, "quux", str)
}

func TestRelationsCodec(t *testing.T) {
	rels := Relations{
		Relation{FromBlockID: 1, ToBlockID: 2, Kind: Contains, Name: "child"},
		Relation{FromBlockID: 2, ToBlockID: 4, Kind: References, Name: "palette"},
		Relation{FromBlockID: 3, ToBlockID: 100, Kind: Contains, Name: "metadata"},
	}

	buf := bytes.Buffer{}
	hnd := &relationsHandler{}

	ec := NewEncodeContext()
	n, err := hnd.GoblinEncode(ec, &buf, rels)
	assert.NoError(t, err)
	assert.Equal(t, BlockVersion(1), n)

	assert.True(t, ec.Strings.Has("child"))
	assert.True(t, ec.Strings.Has("palette"))
	assert.True(t, ec.Strings.Has("metadata"))

	reloaded, err := hnd.GoblinDecode(NewDecodeContext(&buf, ec.Strings), &buf, 1, 48)
	assert.NoError(t, err)
	assert.Equal(t, rels, reloaded.(Relations))
}
