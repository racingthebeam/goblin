package goblin

import "fmt"

type Block struct {
	ID   BlockID
	Type BlockType
	Name string
	Data any

	C *Container
}

func (b *Block) Children() ([]*Block, error) {
	ch := b.C.Relations().ChildrenOf(nil, b.ID)
	out := make([]*Block, 0, len(ch))

	for _, c := range ch {
		cb, ok := b.C.Block(c.ToBlockID)
		if !ok {
			return nil, fmt.Errorf("missing child ID %d", c)
		}
		out = append(out, cb)
	}

	return out, nil
}

func (b *Block) FirstChildWithType(t BlockType) *Block {
	return nil
}

func (b *Block) ChildrenWithType(t BlockType) []*Block {
	return nil
}
