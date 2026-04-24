package main

import (
	"github.com/alecthomas/kong"
	"github.com/racingthebeam/goblin/cli"
)

type CLI struct {
	Goblin cli.CLI `embed:""`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli)
	ctx.FatalIfErrorf(ctx.Run())
}
