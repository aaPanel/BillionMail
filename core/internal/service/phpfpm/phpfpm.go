package phpfpm

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/gconv"
	"io"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/tomasen/fcgi_client"
)

// PHPFpmHandlerConfig holds the configuration for the PHP-FPM handler.
type PHPFpmHandlerConfig struct {
	Network string // Network type (tcp, unix)
	Addr    string // Address of the FastCGI server
	Root    string // Document root for PHP files
	Static  string // Serve static files directly
}

// PHPFpm Handler Factory creates a handler function for PHP-FPM.
func PHPFpmHandlerFactory(config PHPFpmHandlerConfig) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		// Get the requested file path
		filePath := "/" + r.Get("any").String()

		// Check if the file path is empty
		if filePath == "/" {
			filePath = "/index.php"
		}

		if !strings.HasSuffix(filePath, ".php") {
			r.Response.ServeFile(filepath.Join(config.Static, filePath))
			return
		}

		https := "off"

		if r.GetSchema() == "https" {
			https = "on"
		}

		// Get the remote address and port
		remoteAddr, port, _ := net.SplitHostPort(r.RemoteAddr)

		// Create environment variables for FastCGI
		env := map[string]string{
			"SCRIPT_FILENAME": filepath.Join(config.Root, filePath), // Adjust document root as needed
			"REQUEST_METHOD":  r.Method,
			"SCRIPT_NAME":     filePath,
			"REQUEST_URI":     r.RequestURI,
			"QUERY_STRING":    r.URL.RawQuery,
			"CONTENT_TYPE":    r.Header.Get("Content-Type"),
			"CONTENT_LENGTH":  r.Header.Get("Content-Length"),
			"REMOTE_ADDR":     remoteAddr,
			"SERVER_NAME":     r.Host,
			"SERVER_PORT":     port,
			"SERVER_PROTOCOL": r.Proto,
			"HTTPS":           https,
			"REQUEST_TIME":    gconv.String(time.Now().Unix()),
			"HTTP_HOST":       r.Host,
			"HTTP_USER_AGENT": r.Header.Get("User-Agent"),
			"HTTP_ACCEPT":     r.Header.Get("Accept"),
		}

		// Add HTTP headers to environment
		for key, values := range r.Header {
			if len(values) > 0 {
				env["HTTP_"+key] = values[0]
			}
		}

		fc, err := fcgiclient.Dial(config.Network, config.Addr)

		if err != nil {
			glog.Error(context.Background(), err)
			return
		}

		defer fc.Close()

		resp, err := fc.Request(env, r.Body)

		if err != nil {
			glog.Error(context.Background(), err)
			return
		}

		defer resp.Body.Close()

		// Copy headers from PHP response to our response
		for key, values := range resp.Header {
			for _, value := range values {
				r.Response.Header().Add(key, value)
			}
		}

		// Set status code
		r.Response.WriteStatus(resp.StatusCode)

		// Copy response body
		_, err = io.Copy(r.Response.Writer, resp.Body)

		if err != nil {
			glog.Error(context.Background(), err)
			return
		}
	}
}
