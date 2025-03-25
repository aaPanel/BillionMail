package main

import (
	_ "billionmail-core/internal/packed"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"billionmail-core/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
