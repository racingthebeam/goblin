package goblin

func init() {
	RegisterBlockType(BlockTypeRelations, &relationsHandler{})
	RegisterBlockType(BlockTypeStrings, &stringsHandler{})
	RegisterBlockType(BlockTypeFileInfo, &fileInfoHandler{})
	RegisterBlockType(BlockTypeBlob, &blobHandler{})
}
