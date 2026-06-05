# goblin

Goblin is a generic block-based container format for binary data with support for string interning, block relationship modelling and per-block compression. It was extracted from the [BEAM256 Low-Level Fantasy Console project](https://github.com/racingthebeam/beam256).

## Table of Contents

  - Concepts
  - Built-in Block Types
  - (Go) CLI Tool Usage

## Concepts

  - Container -
  - Block -
  - Block ID - a 32 bit unsigned integer, greater than zero, that uniquely identifies each block in a container. Block IDs may be manually allocated, or auto-generated when a block is attached. Whether or not block IDs are meaningful is dependent on the project - simple projects may stick to a set of well-known IDs, whereas larger projects with dynamic data may need to be more generic. A handful of block IDs (>= `0xFFFF0000`) are reserved for internal use (e.g. string and relation tables).
  - Block type - again, a 32 bit unsigned integer, greater than zero, that specifies the type of each block. The block type is directly used to map on-disk blocks to their corresponding `BlockTypeHandler` instances, via a `Registry`. The block type's MSB specifies whether the block type is **public** (MSB set) or **private** (MSB unset). Private block types are intended for internal use by users, without expectation that they will not clash with other users' private block types. For use cases requiring public interop, see [Public Block Types](PUBLIC_BLOCK_TYPES.md))

## Custom Blocks

Support for custom block types is implemented by creating a handler conforming to the `BlockTypeHandler` interface and then registering it with a `Registry` (either implicit or explicit).

### `BlockTypeHandler` Implementation

```
type BlockTypeHandler interface {
	GoblinName() string
	GoblinDump(w io.Writer, b any, opts *DumpOpts) error
	GoblinLint(c any) error
	GoblinCompression(version BlockVersion) BlockCompression
	GoblinEncode(dst *EncodeContext, w io.Writer, c any) (BlockVersion, error)
	GoblinDecode(src *DecodeContext, r io.Reader, ver BlockVersion, size int64) (any, error)
}
```

#### `GoblinName() string`

Returns the name of this block type, used for diagnostics only.

By convention, built-in block types have `UPPERCASE` names and all others are `lowercase`.

#### `GoblinDump(w io.Writer, b any, opts *DumpOpts) error`

Dump the block contents `b` to output `w`, for diagnostics/inspection purposes. `opts` specifies the desired verbosity (summary/preview/full), and whether output should be colorized.

#### `GolinLint(b any) error`

Check block data `b` for validity.

#### `GoblinCompression() BlockCompression`

Returns the desired compression setting (`NoCompression`, `GZip`, `ZLib`) to be employed when encoding new block data.

#### `GoblinEncode(dst *EncodeContext, w io.Writer, b any) (BlockVersion, error)`

Encode block data `b` to `w`, returning a version number describing the encoded format.

If the block includes string data, these may be interned by using `dst.Strings.Add()`.

#### `GoblinDecode(src *DecodeContext, r io.Reader, v BlockVersion, size int64) (any, error)`

Decode block data from `r` and return it as a fully hydrated object. Version `v` is that which is stored in the on-disk block index - the decoder must inspect this and select the appropriate decode strategy.

`size` is the full, uncompressed size of the block's data, and `r` is limited to reading this number of bytes so there is no risk of overshooting the block bounds.

If the block includes interned string data, decode this using `src.Strings.Lookup()`.

### Block Handler Registration

To register a block type with the default/implicit registry, use `goblin.RegisterBlockType()`.

This should be suitable for most use-cases, and is required when using Goblin's built-in CLI functionality.

```golang
func init() {
    myHandler := &MyCoolBlockTypeHandler{}
    goblin.RegisterBlockType(myBlockType, myHandler)
}
```

For more complex scenarios it may be necessary to use a custom `Registry`:

```golang
func customRegistryExample() {
    // Create a new block type registry
    // The registry is automatically populated with Goblin's built-in block types
    reg := goblin.NewRegistry()

    reg.RegisterBlockType(myBlockType, &MyCoolBlockTypeHandler{})
}
```

When using custom regstries, use `registry.NewEncoder()`/`registry.NewDecoder()` to create encoders/decoders that are preconfigured for use with the given registry. Alternatively, the `WithRegistry()` option can also be passed to the naked encoder/decoder constructors.

## Built-in Block Types

### `FILEINFO`

### `METADATA`

### `RELATIONS`

Models inter-block relationships. Each relationship is modelled as a tuple of (from block ID, to block ID, type, name), where type is one of `Contains` or `References` (the former allowing for the representation of hierarchy), and name is an arbitrary string that describes the relationship; there is no requirement that this be unique.

A Goblin container must only contain one single `RELATIONS` block.

### `STRINGS`

Interned table, mapping integers to strings. Accessible by all blocks during encode and decode phases. In many circumstances using `STRINGS` allows blocks to encode their data using fixed-length records while still permitting arbitrary length strings as descriptors.

A Goblin container must only contain one single `STRINGS` block.

### `BLOB`

Stores a chunk of bytes, no compression. Useful in simple circumstances where you don't need to overhead of defining your own block type.
