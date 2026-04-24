package goblin

import (
	"fmt"
	"sync"
)

var (
	globalRegistry = NewRegistry()
)

func LookupBlockType(bt BlockType) (BlockTypeHandler, bool) {
	return globalRegistry.LookupBlockType(bt)
}

func RegisterBlockType(bt BlockType, hnd BlockTypeHandler) {
	globalRegistry.RegisterBlockType(bt, hnd)
}

type Registry struct {
	blockTypes     map[BlockType]BlockTypeHandler
	blockTypesLock sync.RWMutex
}

func NewRegistry() *Registry {
	r := &Registry{
		blockTypes: map[BlockType]BlockTypeHandler{},
	}

	r.RegisterBlockType(BlockTypeRelations, &relationsHandler{})
	r.RegisterBlockType(BlockTypeStrings, &stringsHandler{})
	r.RegisterBlockType(BlockTypeFileInfo, &fileInfoHandler{})
	r.RegisterBlockType(BlockTypeBlob, &blobHandler{})

	return r
}

func (r *Registry) LookupBlockType(bt BlockType) (BlockTypeHandler, bool) {
	r.blockTypesLock.RLock()
	defer r.blockTypesLock.RUnlock()

	hnd, ok := r.blockTypes[bt]
	return hnd, ok
}

func (r *Registry) RegisterBlockType(bt BlockType, hnd BlockTypeHandler) {
	r.blockTypesLock.Lock()
	defer r.blockTypesLock.Unlock()

	if _, ok := r.blockTypes[bt]; ok {
		panic(fmt.Errorf("duplicate block type %d", bt))
	}

	r.blockTypes[bt] = hnd
}
