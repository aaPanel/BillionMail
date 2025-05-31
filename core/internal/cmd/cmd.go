package cmd

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/controller/abnormal_recipient"
	"billionmail-core/internal/controller/batch_mail"
	"billionmail-core/internal/controller/campaign"
	"billionmail-core/internal/controller/contact"
	"billionmail-core/internal/controller/dockerapi"
	"billionmail-core/internal/controller/domains"
	"billionmail-core/internal/controller/email_template"
	"billionmail-core/internal/controller/files"
	"billionmail-core/internal/controller/languages"
	"billionmail-core/internal/controller/mail_boxes"
	"billionmail-core/internal/controller/mail_services"
	"billionmail-core/internal/controller/overview"
	"billionmail-core/internal/controller/rbac"
	"billionmail-core/internal/controller/relay"
	"billionmail-core/internal/controller/settings"
	"billionmail-core/internal/service/database_initialization"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/maillog_stat"
	"billionmail-core/internal/service/middlewares"
	"billionmail-core/internal/service/phpfpm"
	"billionmail-core/internal/service/public"
	rbac2 "billionmail-core/internal/service/rbac"
	"billionmail-core/internal/service/redis_initialization"
	"billionmail-core/internal/service/rspamd"
	"billionmail-core/internal/service/timers"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/util/gconv"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	Main = gcmd.Command{
		Name:  consts.DEFAULT_SERVER_NAME,
		Usage: consts.DEFAULT_SERVER_NAME,
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			if v := parser.GetOpt("version"); v != nil {
				fmt.Println(fmt.Sprintf("v%s", g.Cfg().MustGet(ctx, "server.version", "0.1").String()))
				return nil
			}

			// Init Database
			err = database_initialization.InitDatabase()

			if err != nil {
				g.Log().Error(ctx, "initialize databases failed ", err)
				return err
			}

			// Init Redis
			err = redis_initialization.InitRedis()

			if err != nil {
				g.Log().Error(ctx, "initialize redis failed ", err)
				return err
			}

			// get safe path
			safepath, _ := public.DockerEnv("SafePath")

			if safepath != "" {
				safepath = strings.TrimPrefix(safepath, "/")
			}

			// Start timers
			err = timers.Start(ctx)

			if err != nil {
				g.Log().Error(ctx, "start timers failed ", err)
				return err
			}

			// Connect to Docker
			dk, err := docker.NewDockerAPI()

			if err != nil {
				g.Log().Error(ctx, "failed to connect to docker-api ", err)
				return err
			}

			defer dk.Close()

			// Init Rspamd worker-controller
			err = rspamd.InitWorkerController()

			if err != nil {
				g.Log().Warning(ctx, "rspamd init failed ", err)
				err = nil
			}

			// Create a new server instance
			s := g.Server(consts.DEFAULT_SERVER_NAME)

			// Use Redis for session storage
			// s.SetSessionStorage(gsession.NewStorageRedis(g.Redis()))

			// Define excluded URIs
			excludesURIs := map[string]struct{}{
				"/favicon.ico":                {},
				"/robots.txt":                 {},
				"/unsubscribe.html":           {},
				"/api/unsubscribe/user_group": {},
				"/api/unsubscribe":            {},
			}

			// Bind Server Hooks
			s.BindHookHandlerByMap("/*", map[ghttp.HookName]ghttp.HandlerFunc{
				ghttp.HookBeforeServe: func(r *ghttp.Request) {
					// Safe path check
					if safepath != "" {
						if r.URL.Path == "/"+safepath {
							// Set session
							err := r.Session.Set("safe_path_pass", true)

							if err != nil {
								g.Log().Error(ctx, "set safe_path_pass failed ", err)
							}

							// r.Response.RedirectTo("/")
							r.SetCtxVar("JustVisitedSafePath", true)
							return
						}

						// check if the request is in the excluded URIs
						if _, ok := excludesURIs[r.URL.Path]; ok {
							return
						}

						if !r.IsFileRequest() && !strings.HasPrefix(r.URL.Path, "/api/") {
							return
						}

						if !r.Session.MustGet("safe_path_pass", false).Bool() {
							if strings.HasPrefix(r.URL.Path, "/api/") {
								resp := public.CodeMap[404]
								resp.Msg = "access denied"
								r.Response.WriteJson(resp)
							} else {
								g.Log().Debug(ctx, "Safe path not passed ", r.URL.Path)
								r.Response.WriteHeader(404)
							}
							r.ExitAll()
							return
						}
					}
				},
				// ghttp.HookAfterServe: func(r *ghttp.Request) {},
				// ghttp.HookBeforeOutput: func(r *ghttp.Request) {},
				// ghttp.HookAfterOutput: func(r *ghttp.Request) {},
			})

			// Register Common Handlers
			s.Group("/", func(group *ghttp.RouterGroup) {
				// Add CORS middleware
				group.Middleware(ghttp.MiddlewareCORS)

				// Add docker client middleware
				group.Middleware(func(r *ghttp.Request) {
					r.SetCtxVar(consts.DEFAULT_DOCKER_CLIENT_CTX_KEY, dk)
					r.Middleware.Next()
				})

				group.Bind(
					campaign.NewV1(),
				)
			})

			// Register Apis
			s.Group("/api", func(group *ghttp.RouterGroup) {
				// Add CORS middleware
				group.Middleware(ghttp.MiddlewareCORS)

				// Add docker client middleware
				group.Middleware(func(r *ghttp.Request) {
					r.SetCtxVar(consts.DEFAULT_DOCKER_CLIENT_CTX_KEY, dk)
					r.Middleware.Next()
				})

				// Add JWT middleware
				group.Middleware(rbac2.JWT().JWTAuthMiddleware)

				// Add RBAC middleware
				// group.Middleware(middlewares.NewRBACMiddleware().PermissionCheck)

				// group.Middleware(ghttp.MiddlewareHandlerResponse)

				// Add response middleware
				group.Middleware(middlewares.HandleApiResponse)

				group.Bind(
					rbac.NewV1(),
					domains.NewV1(),
					mail_boxes.NewV1(),
					overview.NewV1(),
					dockerapi.NewV1(),
					contact.NewV1(),
					email_template.NewV1(),
					batch_mail.NewV1(),
					files.NewV1(),
					abnormal_recipient.NewV1(),
					languages.NewV1(),
					mail_services.NewV1(),
					relay.NewV1(),
					settings.NewV1(),
				)
			})

			// Add PHP-FPM middleware
			s.BindMiddleware("/roundcube/*any", func(r *ghttp.Request) {
				if r.Method == "POST" && strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
					// Get and store the request body
					r.GetBody()
					r.Middleware.Next()
					return
				}
				r.Middleware.Next()
			})

			// Binding PHP-FPM handler
			s.BindHandler("/roundcube/*any", phpfpm.PHPFpmHandlerFactory(phpfpm.PHPFpmHandlerConfig{
				Network: "unix",
				Addr:    consts.PHP_FPM_SOCK_PATH,
				Root:    consts.ROUNDCUBE_ROOT_PATH_IN_CONTAINER,
				Static:  consts.ROUNDCUBE_ROOT_PATH,
			}))

			// Proxy 60880 port for ACME challenge
			ACMEChallengeProxy := httputil.NewSingleHostReverseProxy(&url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:60880",
			})
			ACMEChallengeProxy.Transport = &http.Transport{
				MaxIdleConns:        5,
				MaxIdleConnsPerHost: 1,
				IdleConnTimeout:     5 * time.Second,
			}
			s.BindHandler("/.well-known/acme-challenge/*any", func(r *ghttp.Request) {
				ACMEChallengeProxy.ServeHTTP(r.Response.BufferWriter, r.Request)
			})

			// Proxy Rspamd GUI
			var rspamdProxy *httputil.ReverseProxy
			rspamdProxyMutex := sync.Mutex{}
			s.BindHandler("/rspamd/*any", func(r *ghttp.Request) {
				if !r.Session.MustGet("UserLogin", false).Bool() {
					r.Response.WriteHeader(403)
					r.Response.Write([]byte("Access denied"))
					return
				}

				// Ensure the Rspamd proxy created only once
				if rspamdProxy == nil {
					func() {
						rspamdProxyMutex.Lock()
						defer rspamdProxyMutex.Unlock()

						if rspamdProxy == nil {
							host := "127.0.0.1:21334"

							if public.IsRunningInContainer() {
								host = "rspamd:11334"
							}

							rspamdProxy = httputil.NewSingleHostReverseProxy(&url.URL{
								Scheme: "http",
								Host:   host,
							})

							rspamdProxy.Transport = &http.Transport{
								MaxIdleConns:        10,
								MaxIdleConnsPerHost: 2,
								IdleConnTimeout:     15 * time.Second,
							}
						}
					}()
				}

				var password string
				err = public.OptionsMgrInstance.GetOption(ctx, "rspamd_worker_controller_password", &password)

				if err == nil {
					r.Header.Set("Password", password)
				}

				r.URL.Path = strings.Replace(r.URL.Path, "/rspamd", "", 1)

				if r.URL.Path == "" {
					r.URL.Path = "/"
				}

				rspamdProxy.ServeHTTP(r.Response.BufferWriter, r.Request)
			})

			// Email Campaign Tracker
			s.BindHandler("/pmta/*any", func(r *ghttp.Request) {
				maillog_stat.CampaignEventHandler(r, r.Get("any").String())
			})

			// Add static file handler
			s.BindHandler("/*any", func(r *ghttp.Request) {
				if strings.HasPrefix(r.URL.Path, "/api/") {
					r.Response.WriteHeader(404)
					return
				}

				if r.GetCtxVar("JustVisitedSafePath", false).Bool() {
					r.Response.RedirectTo("/")
					return
				}

				if safepath != "" && !r.Session.MustGet("safe_path_pass", false).Bool() {
					r.Response.WriteHeader(404)
					return
				}

				r.Response.ServeFile("public/dist/index.html")
			})

			// Generate self-signed certificate if not exists
			public.SelfSignedCert().Generate()

			// Enable HTTPS
			s.EnableHTTPS(public.AbsPath(filepath.Join(consts.SSL_PATH, "cert.pem")), public.AbsPath(filepath.Join(consts.SSL_PATH, "key.pem")))

			// attempt add http port
			if httpPort, err := public.DockerEnv("HTTP_PORT"); err == nil && httpPort != "" {
				s.SetPort(gconv.Int(httpPort))
			}

			// Set HTTPS ports
			s.SetHTTPSPort(g.Cfg().MustGet(ctx, "server.httpsPort", 443).Int())
			if httpsPort, err := public.DockerEnv("HTTPS_PORT"); err == nil && httpsPort != "" {
				s.SetHTTPSPort(gconv.Int(httpsPort))
			}

			s.SetSwaggerUITemplate(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>openAPI UI</title>
  </head>
  <body>
    <div id="openapi-ui-container" spec-url="{SwaggerUIDocUrl}" theme="light"></div>
    <script src="https://cdn.jsdelivr.net/npm/openapi-ui-dist@latest/lib/openapi-ui.umd.js"></script>
  </body>
</html>`)

			s.Run()

			return nil
		},
	}
)
