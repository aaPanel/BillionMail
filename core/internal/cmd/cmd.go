package cmd

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/controller/abnormal_recipient"
	"billionmail-core/internal/controller/askai"
	"billionmail-core/internal/controller/batch_mail"
	"billionmail-core/internal/controller/campaign"
	"billionmail-core/internal/controller/contact"
	"billionmail-core/internal/controller/dockerapi"
	"billionmail-core/internal/controller/domains"
	"billionmail-core/internal/controller/email_template"
	"billionmail-core/internal/controller/files"
	fb_ctrl "billionmail-core/internal/controller/frostbyte"
	"billionmail-core/internal/controller/languages"
	"billionmail-core/internal/controller/mail_boxes"
	"billionmail-core/internal/controller/mail_services"
	"billionmail-core/internal/controller/middleware"
	"billionmail-core/internal/controller/operation_log"
	"billionmail-core/internal/controller/overview"
	"billionmail-core/internal/controller/rbac"
	"billionmail-core/internal/controller/relay"
	"billionmail-core/internal/controller/settings"
	"billionmail-core/internal/controller/subscribe_list"
	"billionmail-core/internal/controller/tags"
	video_outreach_ctrl "billionmail-core/internal/controller/video_outreach"
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
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/util/gconv"
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

			containerName := consts.SERVICES.Core
			container, err := dk.GetContainerByName(ctx, containerName)
			if err == nil {
				if container.Labels != nil {
					workingDir, exists := container.Labels["com.docker.compose.project.working_dir"]
					if exists {
						if workingDir != "" {
							public.HostWorkDir = workingDir
						}
					}
				}
			}
			if public.HostWorkDir == "" {
				public.HostWorkDir = public.AbsPath("../")
				g.Log().Warning(ctx, "HostWorkDir not found via Docker label, using fallback: ", public.HostWorkDir)
			}

			// Operation log type
			public.LogTypeMap = consts.GetLogTypeMap()

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

			// ip whitelist middleware
			s.Use(middleware.IPWhitelist)

			// Define excluded URIs
			excludesURIs := map[string]struct{}{
				"/favicon.ico":                   {},
				"/robots.txt":                    {},
				"/unsubscribe.html":              {},
				"/unsubscribe_new.html":          {},
				"/api/unsubscribe/user_group":    {},
				"/api/unsubscribe":               {},
				"/api/unsubscribe_new":           {},
				"/api/batch_mail/api/send":       {},
				"/api/batch_mail/api/batch_send": {},
				"/api/subscribe/confirm":         {},
				"/api/subscribe/submit":          {},
				"/api/languages/get":             {},
				"/already_subscribed.html":       {},
				"/subscribe_confirm.html":        {},
				"/subscribe_form.html":           {},
				"/subscribe_success.html":        {},
				"/unsubscribe_success.html":      {},
				"/subscribe_form_code.html":      {},
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

						if r.IsFileRequest() {
							return
						}

						if !r.Session.MustGet("safe_path_pass", false).Bool() {
							if strings.HasPrefix(r.URL.Path, "/api/") {
								// Check if the request is an API token request
								if claims, err := rbac2.JWT().ParseTokenByRequest(r); err == nil && claims != nil && claims.ApiToken {
									return
								}
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
				group.Middleware(middlewares.NewRBACMiddleware().PermissionCheck)

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
					subscribe_list.NewV1(),
					operation_log.NewV1(),
					askai.NewV1(),
					tags.NewV1(),
					fb_ctrl.NewV1(),
					video_outreach_ctrl.NewV1(),
				)
			})

			// Public video landing page (no auth — recipients click from email)
			s.BindHandler("GET:/landing/video", video_outreach_ctrl.ServeLandingPage)

			// FrostByte API reverse proxy (JWT-protected)
			frostbyteProxy := httputil.NewSingleHostReverseProxy(&url.URL{
				Scheme: "http",
				Host:   "frostbyte-api:8001",
			})
			frostbyteProxy.Transport = &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 2,
				IdleConnTimeout:     30 * time.Second,
			}
			s.Group("/api/frostbyte/proxy", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareCORS)
				group.Middleware(rbac2.JWT().JWTAuthMiddleware)
				group.ALL("/*any", func(r *ghttp.Request) {
					// Strip /api/frostbyte/proxy prefix
					r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/frostbyte/proxy")
					if r.URL.Path == "" {
						r.URL.Path = "/"
					}
					frostbyteProxy.ServeHTTP(r.Response.BufferWriter, r.Request)
				})
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
				if _, err := rbac2.JWT().ParseToken(r.Session.MustGet("SignedToken", "").String()); err != nil {
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
			defaultCrtPair, _ := tls.LoadX509KeyPair(public.AbsPath(consts.SSL_PATH, "cert.pem"), public.AbsPath(consts.SSL_PATH, "key.pem"))
			crtMap := make(map[string]*tls.Certificate)
			crtMapMutex := sync.RWMutex{}
			s.EnableHTTPS(public.AbsPath(consts.SSL_PATH, "cert.pem"), public.AbsPath(consts.SSL_PATH, "key.pem"), &tls.Config{
				GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
					domain := info.ServerName

					crtMapMutex.RLock()
					if crtPair, ok := crtMap[domain]; ok {
						crtMapMutex.RUnlock()
						return crtPair, nil
					}
					crtMapMutex.RUnlock()

					if exists, _ := g.DB().Model("domain").Where("a_record", public.FormatMX(domain)).Exist(); !exists {
						return &defaultCrtPair, nil
					}

					g.Log().Debug(ctx, "Get TLS certificate for domain", domain)

					crtMapMutex.Lock()
					defer crtMapMutex.Unlock()

					if crt, ok := crtMap[domain]; ok {
						return crt, nil
					}

					crtPath := public.AbsPath(consts.SSL_PATH, public.FormatMX(domain), "fullchain.pem")
					keyPath := public.AbsPath(consts.SSL_PATH, public.FormatMX(domain), "privkey.pem")

					if !public.FileExists(crtPath) || !public.FileExists(keyPath) {
						return &defaultCrtPair, nil
					}

					crtPair, err := tls.LoadX509KeyPair(crtPath, keyPath)

					if err != nil {
						g.Log().Warning(ctx, "Failed to load TLS certificate for domain", domain, err)
						return &defaultCrtPair, nil
					}

					crtMap[domain] = &crtPair

					return &crtPair, nil
				},
			})

			// attempt add http port
			if httpPort, err := public.DockerEnv("HTTP_PORT"); err == nil && httpPort != "" {
				s.SetPort(gconv.Int(httpPort))
			}

			// Set HTTPS ports
			s.SetHTTPSPort(g.Cfg().MustGet(ctx, "server.httpsPort", 443).Int())
			if httpsPort, err := public.DockerEnv("HTTPS_PORT"); err == nil && httpsPort != "" {
				s.SetHTTPSPort(gconv.Int(httpsPort))
			}

			var apiDocEnabled bool
			err = public.OptionsMgrInstance.GetOption(ctx, "API_DOC_ENABLED", &apiDocEnabled)
			if err != nil {
				apiDocEnabled = false
			}

			if apiDocEnabled {
				s.SetOpenApiPath("/api.json")
				s.SetSwaggerPath("/swagger")
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
