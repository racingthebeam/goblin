package goblin

type Block struct {
	ID      BlockID
	Type    BlockType
	Version int
	Name    string
	Data    any
	// TODO: encoding
	// TODO: compression

	C *Container
}

func (b *Block) Children() []*Block {
	return nil
}

func (b *Block) FirstChildWithType(t BlockType) *Block {
	return nil
}

func (b *Block) ChildrenWithType(t BlockType) []*Block {
	return nil
}
