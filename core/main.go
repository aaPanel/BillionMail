package main

import (
	"billionmail-core/internal/consts"
	_ "billionmail-core/internal/packed"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"

	"github.com/gogf/gf/v2/os/gctx"

	"billionmail-core/internal/cmd"
	"billionmail-core/internal/service/public"
)

func main() {
	dbPass, err := public.ReadFile(consts.DEFAULT_DB_PASS_FILE)

	if err != nil {
		panic(fmt.Errorf("Read database password failed: %v", err))
	}

	// Setting database configuration
	_ = gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			gdb.ConfigNode{
				Host: "pgsql",
				Port: "5432",
				User: "billionmail",
				Pass: dbPass,
				Name: "billionmail",
				Type: "pgsql",
				Role: "master",
			},
		},
	})

	cmd.Main.Run(gctx.GetInitCtx())
}
