package goblin

type Block struct {
	ID   BlockID   // Block ID
	Type BlockType // Block type
	Name string    // Block name
	Data any       // Block data

	C *Container // Container that owns this block
}

// Valid() returns true if the given block is valid. Use this method to
// check for existence when using Container lookup methods like FirstBlockOfType()
func (b *Block) Valid() bool { return b.ID.Valid() }

func (b *Block) Children() []Block                  { return b.C.Children(b.ID) }
func (b *Block) FirstChildOfType(t BlockType) Block { return b.C.FirstChildOfType(b.ID, t) }
func (b *Block) ChildrenOfType(t BlockType) []Block { return b.C.ChildrenOfType(b.ID, t) }
