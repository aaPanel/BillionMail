package video_outreach

import (
	"html/template"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
)

const landingPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Personalized Video for {{.Name}}</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#0f172a;color:#e2e8f0;min-height:100vh;display:flex;align-items:center;justify-content:center}
.container{max-width:800px;width:100%;padding:2rem}
h1{font-size:1.5rem;font-weight:600;margin-bottom:1.5rem;text-align:center}
.video-wrap{position:relative;border-radius:12px;overflow:hidden;box-shadow:0 25px 50px -12px rgba(0,0,0,0.5);background:#1e293b}
video{width:100%;display:block}
.cta{display:block;margin:2rem auto 0;padding:.875rem 2rem;background:#3b82f6;color:#fff;border:none;border-radius:8px;font-size:1rem;font-weight:500;cursor:pointer;text-align:center;text-decoration:none;transition:background .2s}
.cta:hover{background:#2563eb}
</style>
</head>
<body>
<div class="container">
<h1>A quick message for you, {{.Name}}</h1>
<div class="video-wrap">
<video controls playsinline poster="{{.Thumb}}">
<source src="{{.Video}}" type="video/mp4">
Your browser does not support video.
</video>
</div>
<a class="cta" href="{{.CTA}}">Book a Call</a>
</div>
</body>
</html>`

var landingTmpl = template.Must(template.New("landing").Parse(landingPageHTML))

// ServeLandingPage renders the personalized video landing page.
// Public route — no auth required. Query params: video, thumb, name, cta.
func ServeLandingPage(r *ghttp.Request) {
	video := r.GetQuery("video").String()
	thumb := r.GetQuery("thumb").String()
	name := r.GetQuery("name").String()
	cta := r.GetQuery("cta").String()

	if video == "" || name == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "missing video or name param")
		return
	}
	if cta == "" {
		cta = "#"
	}

	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = landingTmpl.Execute(r.Response.Writer, map[string]string{
		"Video": video,
		"Thumb": thumb,
		"Name":  name,
		"CTA":   cta,
	})
}
