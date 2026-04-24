package cli

type CLI struct {
	Dump    CmdDump    `cmd:"" name:"dump"`
	Strings CmdStrings `cmd:"" name:"strings"`
}
