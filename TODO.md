# Metadata etc.

Here's my current thoughts on this:

  - every block has optional metadata
  - metadata is key/value pairs:
    - key - string, interned via string table
    - type tag (int32, int64, float32, float64, boolean, string)
    - value - encoded as per type

In the future, arrays could be encoded by setting the high bit on the type tag then encoding a length field.
Nested objects could be another type tag.
(but we'll leave the composite things out for now)

For the Go implementation:

  - block encode - returns a metadata map
  - block decode - receives a metadata map

For JavaScript - decode to native types.
If we do a JS encoder, it can return an auxiliary lookup for types to specify encoding

This simplifies `FILEINFO` too - `FILEINFO` is just a block that is metadata only. At the container level, it's just a `map[string]any`, and we write it as a `FILEINFO` block at encode time.

Future feature: images

  - come up with a block convention for storing images (PNG, JPEG, BMP, ICO etc)
  - metadata: MIME type, width, height, possibly more...

Then we build a couple of block types on top of this:

  - `ICON`; block IDs - `FILEICON`
  - `IMAGE`; block IDs - `FILEIMAGESUMMARY`

[go-ico](https://pkg.go.dev/github.com/sergeymakinen/go-ico#Decode)
^ looks useful - notice that it can return a list of images.
If we use this, it must not be part of the public interface - just use the stdlib Image functionality.

Key thing - BLOCK TYPES DO NOT IMPLY SEMANTIC MEANING IN THE CONTEXT OF THE FILE.
String table, relations, and fileinfo are sort-of exceptions - but they're more "meta structure".

Other stuff we could have:

  - generic DIRECTORY and FILE types for representing file trees
    - this opens a general question about blocks containing other blocks; natural way to model this is containment at the block data level, but this isn't how Goblin works (relationships are purely a container level feature)
      - maybe we elevate "filesystem" to a top-level block type and the container can have functionality for dealing with this?
  - INDEX for assigning context-sensitive meaning to blocks (necessary?)

## The BEAM256 story:

Maybe worth doing a thought experiment where we think about all the container types we might conceivably have and work out what the model looks like. For example, I'm thinking that there could be "project" file type that can optionally include everything - source files, project metadata, output image, debug symbols. Do we just use duck-typing here? e.g. open a goblin file and see if there's a specific block type, and if it's there, use it?

# TODO

  - [ ] Document meaning of block name
  - [x] Implement string-table support for read/write record
  - [x] Test read/write record
  - [ ] Consider adding block name onto encode/decode context
  - [ ] Make goblin.SimpleHandler generic?
