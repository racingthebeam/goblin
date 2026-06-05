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

func (c *Container) Relations() Relations {
	return c.relations
}

func (c *Container) FirstBlockOfType(typ BlockType) *Block {
	for _, b := range c.blocks {
		if b.Type == typ {
			return b
		}
	}
	return nil
}

func (c *Container) generateBlockID() BlockID {
	for i := BlockID(1); i < BlockReserved; i++ {
		_, exists := c.blocks[i]
		if !exists {
			return i
		}
	}
	panic("failed to generate block ID, file full!")
}
