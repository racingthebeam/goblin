package goblin

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type Container struct {
	blocks    map[BlockID]*Block
	relations Relations
}

func NewContainer() *Container {
	return &Container{
		blocks:    map[BlockID]*Block{},
		relations: Relations{},
	}
}

// BlockIDs returns a sorted list of all block IDs in this container.
func (c *Container) BlockIDs() []BlockID {
	return slices.Sorted(maps.Keys(c.blocks))
}

// Block returns the Block associated with the given ID, or nil if no such block exists.
func (c *Container) Block(id BlockID) (*Block, bool) {
	b, ok := c.blocks[id]
	return b, ok
}

// BlockData returns the data for the specified block ID, or false if there
// is no such block. If et is non-zero, an additional check will be performed
// to ensure the block's type matches the expected.
func (c *Container) BlockData(id BlockID, et BlockType) (any, bool) {
	b, ok := c.blocks[id]
	if !ok || (et != 0 && et != b.Type) {
		return nil, false
	}
	return b.Data, true
}

func (c *Container) SetBlock(id BlockID, typ BlockType, name string, data any) (BlockID, error) {
	if typ == 0 {
		return 0, errors.New("block type must not be zero")
	}

	if id == 0 {
		id = c.generateBlockID()
	}

	if _, exists := c.blocks[id]; exists {
		return 0, fmt.Errorf("block ID %d already exists", id)
	}

	c.blocks[id] = &Block{
		ID:   id,
		Type: typ,
		Name: name,
		Data: data,

		C: c,
	}

	return id, nil
}

// Returns the table of relations. Returned value must not be mutated.
func (c *Container) Relations() Relations {
	return c.relations
}

// Returns the first block with the given type. Check existence with block.Valid().
func (c *Container) FirstBlockOfType(typ BlockType) Block {
	for _, b := range c.blocks {
		if b.Type == typ {
			return *b
		}
	}
	return Block{}
}

// Returns a list of the given block's children; that is, all related blocks
// whose relation type is Contains.
func (c *Container) Children(id BlockID) []Block {
	out := make([]Block, 0)
	for i := range c.relations {
		r := &c.relations[i]
		if r.FromBlockID == id && r.Kind == Contains {
			b := c.blocks[r.ToBlockID]
			if b != nil {
				out = append(out, *c.blocks[r.ToBlockID])
			}
		}
	}
	return out
}

// Returns the first encountered child block with a given type.
func (c *Container) FirstChildOfType(id BlockID, t BlockType) Block {
	for i := range c.relations {
		r := &c.relations[i]
		if r.FromBlockID == id && r.Kind == Contains {
			b := c.blocks[r.ToBlockID]
			if b != nil && b.Type == t {
				return *b
			}
		}
	}
	return Block{}
}

// Returns all child blocks with a given type.
func (c *Container) ChildrenOfType(id BlockID, t BlockType) []Block {
	out := make([]Block, 0)
	for i := range c.relations {
		r := &c.relations[i]
		if r.FromBlockID == id && r.Kind == Contains {
			b := c.blocks[r.ToBlockID]
			if b != nil && b.Type == t {
				out = append(out, *b)
			}
		}
	}
	return out
}

func (c *Container) generateBlockID() BlockID {
	for i := BlockID(1); i < reservedBlockIDBase; i++ {
		_, exists := c.blocks[i]
		if !exists {
			return i
		}
	}
	panic("failed to generate block ID, file full!")
}
