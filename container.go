package goblin

import "io"

type Container struct {
	blocks map[BlockID]*Block

	strings   *Strings
	relations Relations
}

func (c *Container) Dump(w io.Writer) {

}

func (c *Container) Strings() *Strings {
	if c.strings == nil {
		ss := c.FirstBlockOfType(BlockTypeStrings)
		if ss != nil {
			c.strings = ss.Data.(*Strings)
		}
	}
	return c.strings
}

func (c *Container) Relations() Relations {
	if c.relations == nil {
		rs := c.FirstBlockOfType(BlockTypeRelations)
		if rs != nil {
			c.relations = rs.Data.(Relations)
		}
	}
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
