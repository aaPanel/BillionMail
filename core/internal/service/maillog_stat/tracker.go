package maillog_stat

import (
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/contact_activity"
	"billionmail-core/internal/service/frostbyte"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"image"
	"image/color"
	"image/png"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// CampaignEventHandler handle email campaign events
func CampaignEventHandler(r *ghttp.Request, encStr string) {
	ctx := r.GetCtx()
	data := struct {
		MessageId  string `json:"message_id" v:"required|min-length:1"`
		Recipient  string `json:"recipient" v:"required|email"`
		Type       string `json:"type" v:"required|in:open,click"`
		CampaignId int    `json:"campaign_id" v:"required|min:1"`
		Url        string `json:"url" v:"url"`
	}{}
	err := Decrypt(encStr, &data)
	if err != nil {
		r.Response.Write("Invalid -1")
		return
	}

	err = g.Validator().Data(data).Run(ctx)

	if err != nil {
		g.Log().Warning(ctx, "Validate campaign event data failed: ", err)
		r.Response.Write("Invalid -2")
		return
	}

	curTimeMillis := time.Now().UnixMilli()

	postfixMessageID, err := SearchPostfixMessageIdByMessageId(data.MessageId)

	if err != nil {
		g.Log().Warning(ctx, "SearchPostfixMessageIdByMessageId failed: ", err)
	}

	switch data.Type {
	case "open":
		_, err = g.DB().Model("mailstat_opened").Insert(g.Map{
			"campaign_id":        data.CampaignId,
			"log_time_millis":    curTimeMillis,
			"recipient":          data.Recipient,
			"message_id":         data.MessageId,
			"postfix_message_id": postfixMessageID,
		})

		if err != nil {
			g.Log().Error(ctx, err)
		}

		// Update contact activity when email is opened
		var groupId int
		if data.CampaignId > 1000000000 {
			// It's from API
			groupId = 0
		} else {
			var tasks *entity.EmailTask
			err = g.DB().Model("email_tasks").Where("id", data.CampaignId).Scan(&tasks)
			if err != nil {
				g.Log().Error(ctx, "Failed to get email task info: ", err)
				groupId = 0
			} else {
				groupId = tasks.GroupId
			}
		}
		// Update contact activity when email is opened (user interaction)
		contact_activity.UpdateActivityByEmailAndGroup(data.Recipient, groupId)

		// Dispatch to FrostByte
		frostbyte.DispatchOpen(ctx, data.MessageId, data.Recipient)

		// Create a 1x1 transparent PNG image
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		// Set the pixel to transparent
		img.Set(0, 0, color.Transparent)
		r.Response.Header().Set("Content-Type", "image/png")

		err = png.Encode(r.Response.BufferWriter, img)

		if err != nil {
			g.Log().Error(ctx, "Failed to encode PNG: ", err)
		}

		return
	case "click":
		_, err = g.DB().Model("mailstat_clicked").Insert(g.Map{
			"campaign_id":        data.CampaignId,
			"log_time_millis":    curTimeMillis,
			"recipient":          data.Recipient,
			"message_id":         data.MessageId,
			"postfix_message_id": postfixMessageID,
		})

		if err != nil {
			g.Log().Error(ctx, err)
		}
		g.Log().Debug(ctx, "记录点击 !!! Click event data: ", data.Recipient)

		// Update contact activity when link is clicked
		var groupId int
		if data.CampaignId > 1000000000 {
			// It's from API
			groupId = 0
		} else {
			var tasks *entity.EmailTask
			err = g.DB().Model("email_tasks").Where("id", data.CampaignId).Scan(&tasks)
			if err != nil {
				g.Log().Error(ctx, "Failed to get email task info: ", err)
				groupId = 0
			} else {
				groupId = tasks.GroupId
			}
		}
		g.Log().Debug(ctx, "开始记录活跃时间 !!! Click event data: ", data.Recipient, "组->", groupId)
		// Update contact activity when link is clicked (user interaction)
		contact_activity.UpdateActivityByEmailAndGroup(data.Recipient, groupId)

		// Dispatch to FrostByte
		frostbyte.DispatchClick(ctx, data.MessageId, data.Recipient)

		// Redirect to the target URL
		r.Response.RedirectTo(data.Url)
		return
	}

	r.Response.Write("Success")
}

type MailTracker struct {
	originalMailHTML string
	modified         bool
	mailHTML         string
	campaignID       int
	messageID        string
	recipient        string
	baseURL          string
	hrefPattern      *regexp.Regexp
}

func NewMailTracker(mailHTML string, campaignID int, messageID, recipient, baseURL string) *MailTracker {
	return &MailTracker{
		originalMailHTML: mailHTML,
		modified:         false,
		mailHTML:         mailHTML,
		campaignID:       campaignID,
		messageID:        strings.Trim(messageID, "<>"),
		recipient:        recipient,
		baseURL:          fmt.Sprintf("%s/pmta", strings.TrimRight(baseURL, "/")),
		hrefPattern:      regexp.MustCompile(`href\s*=\s*"([^"]+)"`),
	}
}

// TrackLinks handles the tracking of links in the email HTML
func (t *MailTracker) TrackLinks() {
	t.mailHTML = t.hrefPattern.ReplaceAllStringFunc(t.mailHTML, func(s string) string {
		matches := t.hrefPattern.FindStringSubmatch(s)
		if len(matches) < 2 {
			return s
		}

		// Validate URL
		u, err := url.Parse(matches[1])
		if err != nil || u.Scheme == "" || u.Host == "" {
			return s
		}

		return fmt.Sprintf(`href="%s"`, t.GetTrackingURL(matches[1]))
	})

	t.modified = t.originalMailHTML != t.mailHTML
}

func (t *MailTracker) GetTrackingURL(url string) string {
	data := map[string]interface{}{
		"type":        "click",
		"campaign_id": t.campaignID,
		"recipient":   t.recipient,
		"message_id":  t.messageID,
		"url":         url,
	}

	return fmt.Sprintf("%s/%s", t.baseURL, Encrypt(data))
}

// AppendTrackingPixel appends a tracking pixel to the email HTML
func (t *MailTracker) AppendTrackingPixel() {
	trackingPixel := fmt.Sprintf(`<img src="%s" style="display:none" alt="" />`, t.GetTrackingPixel())

	if strings.Contains(t.mailHTML, "</body>") {
		t.mailHTML = strings.Replace(t.mailHTML, "</body>", trackingPixel+"</body>", 1)
		t.modified = true
		return
	}

	if strings.Contains(t.mailHTML, "</html>") {
		t.mailHTML = strings.Replace(t.mailHTML, "</html>", trackingPixel+"</html>", 1)
		t.modified = true
		return
	}

	t.mailHTML += trackingPixel
	t.modified = true
}

// GetTrackingPixel returns the tracking pixel URL
func (t *MailTracker) GetTrackingPixel() string {
	data := map[string]interface{}{
		"type":        "open",
		"campaign_id": t.campaignID,
		"recipient":   t.recipient,
		"message_id":  t.messageID,
	}

	return fmt.Sprintf("%s/%s", t.baseURL, Encrypt(data))
}

// IsModified checks if the HTML has been modified
func (t *MailTracker) IsModified() bool {
	return t.modified
}

// GetOriginalHTML returns the original HTML content
func (t *MailTracker) GetOriginalHTML() string {
	return t.originalMailHTML
}

// GetHTML returns the modified HTML content
func (t *MailTracker) GetHTML() string {
	return t.mailHTML
}
