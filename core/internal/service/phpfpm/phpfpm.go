package phpfpm

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
	"io"

	"github.com/tomasen/fcgi_client"
)

// Transfer PHP-FPM requests to PHP-FPM server via FastCGI
func PHPFpm(r *ghttp.Request) {
	// Get the requested file path
	filePath := r.URL.Path

	// Configure FastCGI client
	fcgiAddr := "/var/run/php/php-fpm.sock" // Default PHP-FPM address, adjust as needed

	// Create environment variables for FastCGI
	env := map[string]string{
		"SCRIPT_FILENAME": "/var/www/html" + filePath, // Adjust document root as needed
		"REQUEST_METHOD":  r.Method,
		"SCRIPT_NAME":     filePath,
		"REQUEST_URI":     r.RequestURI,
		"QUERY_STRING":    r.URL.RawQuery,
		"CONTENT_TYPE":    r.Header.Get("Content-Type"),
		"CONTENT_LENGTH":  r.Header.Get("Content-Length"),
		"REMOTE_ADDR":     r.RemoteAddr,
		"SERVER_NAME":     r.Host,
		"SERVER_PORT":     r.Request.Header.Get("X-Forwarded-Port"),
		"SERVER_PROTOCOL": r.Proto,
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

	fc, err := fcgiclient.Dial("unix", fcgiAddr)

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
