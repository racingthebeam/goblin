# goblin

This repository contains a description of the Goblin file format along with a reference implementation written in Go.

Goblin is a generic block-based container format for binary data with support for string interning, block relationship modelling and configurable compression settings. It was extracted from the [BEAM256 Low-Level Fantasy Console project](https://github.com/racingthebeam/beam256).

Consider using Goblin when you need to store (potentially inter-related) chunks of structured binary data and don't want to think about high-level encoding concerns - blocks are simply Go values that know how to serialize themselves. Additionally, Goblin's built-in string interning simplifies many scenarios, changing otherwise variable-length data to fixed-length values.

**Non-Goals**

  - While Goblin describes a common encoding format for binary data, it is not intended to be a universal self-describing interchange format. That is, in the absence of additional context or cooperation, there is no implicit assumption that program 'A' should be able to meaningfully interpret a Goblin file produced by an unrelated program 'B' (with ith proper coordination this is of course possible, and Goblin supports both public and private identifiers to facilitate this).
  - Goblin is not designed for streaming data, and is intended mainly for payloads that fit comfortably in memory

Why the name "Goblin"? This project was extracted from a custom toolchain, and ELF and DWARF were both taken :).

## Table of Contents

  - Concepts
  - Example Usage
  - Built-in Block Types
  - (Go) CLI Tool Usage

## Concepts

A **Container** is the top-level Goblin entity, containing a list of **Blocks**, each having its own **Block ID** and **Block Type**. Containers also have an optional top-level **File Type** that may be used by implementers to denote the expected structure of the full Container. Containers may exist either on-disk (for storage), or in-memory (for manipulation using this library).

  - Container - 
  - Block -
  - Block ID - a 32 bit unsigned integer, greater than zero, that uniquely identifies each block in a container. Block IDs may be manually allocated, or auto-generated when a block is attached. Whether or not block IDs are meaningful is dependent on the project - simple projects may stick to a set of well-known IDs, whereas larger projects with dynamic data may need to be more generic. A handful of block IDs (>= `0xFFFF0000`) are reserved for internal use (e.g. string and relation tables).
  - Block type - again, a 32 bit unsigned integer, greater than zero, that specifies the type of each block. The block type is directly used to map on-disk blocks to their corresponding `BlockTypeHandler` instances, via a `Registry`. The block type's MSB specifies whether the block type is **public** (MSB set) or **private** (MSB unset). Private block types are intended for internal use by users, without expectation that they will not clash with other users' private block types. For use cases requiring public interop, see [Public Types](PUBLIC_TYPES.md))

## Basic Usage

```golang
const (
    // Each block type has its own ID
    // For private/internal use, you are free to allocate your own values < 0x80000000
    blockTypeA = 1
    blockTypeB = 2
)

func main() {
    // This is the data we want to store in the container - arbitrary Go types
    // Each type has its own registered BlockTypeHandler that is responsible
    // for encoding/decoding (see Custom Blocks, below).
    data1 := &MyDataObjectA{}
    data2 := &MyDataObjectB{}

    // Create a new container and insert the blocks
    // Specify block ID as 0 so the container auto-generates a valid value
    c := goblin.NewContainer()
    id1, _ := c.SetBlock(0, blockTypeA, "data_a", data1)
    id2, _ := c.SetBlock(0, blockTypeB, "data_b", data2)
    log.Printf("Inserted block IDs %d and %d", id1, id2)

    // data2 is a logical child of data1, so insert a relationship
    c.AddRelation(data1, data2, goblin.Contains, "child")

    // Encode the container to a file
    fw, _ := os.Create("out.goblin")
    NewEncoder(fw).Encode(c)
    fw.Close()

    // Now read it back from the same file into a new instance
    fr, _ := os.Open("out.goblin")
    newContainer, _ := NewDecoder(fh).Decode()

    // Fetch a block directly by ID...
    block1 := newContainer.Block(id1)

    // or fetch its data, with optional type guard:
    block1Data, _ := newContainer.BlockData(id1, blockTypeA)
    log.Printf("%s", block1Data.(*MyDataObjectA).SomeField)

    // or look up a block up by type, regardless of ID
    block2 := newContainer.FirstBlockOfType(blockTypeB)
}
```

## Custom Blocks

Support for custom block types is implemented by creating a handler conforming to the `BlockTypeHandler` interface and then registering it with a `Registry` (either implicit or explicit).

**Note:** block types manipulated by `BlockTypeHandler` instances should be reference types (struct pointers, or maps/slices).

### `BlockTypeHandler` Implementation

```
type BlockTypeHandler interface {
	GoblinName() string
	GoblinDump(w io.Writer, b any, opts *DumpOpts) error
	GoblinValidate(b any) error
	GoblinCompression() (BlockCompression, int)
	GoblinEncode(dst *EncodeContext, w io.Writer, b any) (BlockVersion, error)
	GoblinDecode(src *DecodeContext, r io.Reader, ver BlockVersion, size int64) (any, error)
}
```

#### `GoblinName() string`

Returns the name of this block type, used for diagnostics only.

By convention, built-in block types have `UPPERCASE` names and all others are `lowercase`.

#### `GoblinDump(w io.Writer, b any, opts *DumpOpts) error`

Dump the block contents `b` to output `w`, for diagnostics/inspection purposes. `opts` specifies the desired verbosity (summary/preview/full), and whether output should be colorized.

#### `GolinValidate(b any) error`

Check block data `b` for validity prior to writing.

#### `GoblinCompression() (BlockCompression, int)`

Returns the desired compression setting (`NoCompression`, `GZip`, `ZLib`) to be employed when encoding new block data. The second return value is the compression level, e.g. `flate.DefaultCompression` (ignored when compression is disabled).

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

A Goblin container must only contain one single `RELATIONS` block - typically managed automatically by `Container`.

### `STRINGS`

Interned table, mapping integers to strings. Accessible by all blocks during encode and decode phases. In many circumstances using `STRINGS` allows blocks to encode their data using fixed-length records while still permitting arbitrary length strings as descriptors.

A Goblin container must only contain one single `STRINGS` block - typically managed automatically by the encode/decode process.

