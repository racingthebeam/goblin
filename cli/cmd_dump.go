package cli

import (
	"fmt"
	"os"

	"github.com/racingthebeam/goblin"
)

type CmdDump struct {
	File *os.File `arg:""`
}

func (c *CmdDump) Run() error {
	dec := goblin.NewDecoder(c.File)

	dc, err := dec.DecodeHeader()
	if err != nil {
		return err
	}

	con, err := dec.DecodeBlocks(dc)
	if err != nil {
		return err
	}

	fmt.Printf("Index:\n")
	for i := range dc.Index {
		ent := &dc.Index[i]
		name, _ := dc.Strings.Lookup(ent.Name)
		hnd, ok := goblin.LookupBlockType(dc.Index[i].Type)
		typeName := "unknown"
		if ok {
			typeName = hnd.GoblinName()
		}

		fmt.Printf("0x%08X 0x%08X (%16s) %16s %4d %s %8d %8d %8d\n",
			ent.ID,
			ent.Type,
			typeName,
			name,
			ent.Version,
			ent.Compression,
			ent.Offset,
			ent.DataSize,
			ent.CompressedSize,
		)
	}

	fmt.Printf("\n")

	fmt.Printf("Strings:\n")
	strings := dc.Strings.All()
	for i := range strings {
		fmt.Printf("%d: %s\n", i+1, strings[i])
	}

	fmt.Printf("\n")

	fmt.Printf("Relations:\n")
	for _, r := range dc.Relations {
		fmt.Printf("%d -> %d (%s %s)", r.FromBlockID, r.ToBlockID, r.Kind, r.Name)
	}

	fmt.Printf("\n")

	for _, id := range con.BlockIDs() {
		blk, _ := con.Block(id)
		fmt.Printf("Block ID %d - %s - (type: 0x%08x)\n", id, blk.Name, blk.Type)
		hnd, ok := goblin.LookupBlockType(blk.Type)
		if !ok {
			fmt.Printf("UNKNOWN BLOCK\n")
		} else {
			hnd.GoblinDump(os.Stdout, blk.Data, &goblin.DumpOpts{
				Color:   true,
				Verbose: goblin.DumpPreview,
			})
		}
		fmt.Printf("\n")
	}

	return nil
}
