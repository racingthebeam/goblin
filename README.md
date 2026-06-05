# goblin

## Concepts

  - Block ID - a 32 bit unsigned integer, greater than zero, that uniquely identifies each block in a container. Block IDs may be manually allocated, or auto-generated when a block is attached. Whether or not block IDs are meaningful is dependent on the project - simple projects may stick to a set of well-known IDs, whereas larger projects with dynamic data may need to be more generic. A handful of block IDs (>= `0xFFFF0000`) are reserved for internal use (e.g. string and relation tables).
  - Block type - again, a 32 bit unsigned integer, greater than zero, that specifies the type of each block. The block type is directly used to map on-disk blocks to their corresponding `BlockTypeHandler` instances, via a `Registry`. The block type's MSB specifies whether the block type is **public** (MSB set) or **private** (MSB unset). Private block types are intended for internal use by users, without expectation that they will not clash with other users' private block types. For use cases requiring public interop, see [Public Block Types](PUBLIC_BLOCK_TYPES.md))

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
