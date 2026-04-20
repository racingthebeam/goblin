package goblin

import (
	"fmt"
	"sync"
)

var (
	blockTypes     = map[BlockType]BlockTypeHandler{}
	blockTypesLock sync.RWMutex
)

func LookupBlockType(bt BlockType) (BlockTypeHandler, bool) {
	blockTypesLock.RLock()
	defer blockTypesLock.RUnlock()

	hnd, ok := blockTypes[bt]
	return hnd, ok
}

func RegisterBlockType(bt BlockType, hnd BlockTypeHandler) {
	blockTypesLock.Lock()
	defer blockTypesLock.Unlock()

	if _, ok := blockTypes[bt]; ok {
		panic(fmt.Errorf("duplicate block type %d", bt))
	}

	blockTypes[bt] = hnd
}
