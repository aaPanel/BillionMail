package cmd

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/controller/hello"
	"billionmail-core/internal/service/phpfpm"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  consts.DEFAULT_SERVER_NAME,
		Usage: consts.DEFAULT_SERVER_NAME,
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server(consts.DEFAULT_SERVER_NAME)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.NewV1(),
				)
			})

			// Binding PHP-FPM handler
			s.BindHandler("/roundcube/*any", phpfpm.PHPFpmHandlerFactory(phpfpm.PHPFpmHandlerConfig{
				Network: "unix",
				Addr:    "../php-sock/php-fpm.sock",
				Root:    "/var/www/html",
				Static:  "../webmail-data",
			}))
			s.Run()
			return nil
		},
	}
)
