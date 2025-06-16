package email_template

import (
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"regexp"
	"strings"

	"billionmail-core/api/email_template/v1"
)

type ScoreItem struct {
	Name   string  `json:"name"`
	Detail string  `json:"detail"`
	Score  float64 `json:"score"`
}

var spamWords = []string{
	// Forbidden words in Chinese
	"免费", "中奖", "现金", "优惠", "点击这里", "立即领取", "限时特惠", "最后机会", "免费试用",
	"恭喜您", "大奖", "赢取", "抽奖", "红包", "免单", "清仓", "秒杀", "特价", "折扣",
	"投资回报", "高收益", "稳赚不赔", "快速赚钱", "轻松赚钱", "暴利", "一本万利",

	// Forbidden words in English
	"buy now", "free", "winner", "urgent", "limited time", "click here",
	"act now", "apply now", "call now", "don't delete", "exclusive deal",
	"for free", "get it now", "limited offer", "new customers only",
	"order now", "special promotion", "time limited", "while supplies last",
	"earn money", "make money", "extra income", "work from home",
}

var shortLinkDomains = []string{
	// International short chain
	"bit.ly", "tinyurl.com", "goo.gl", "ow.ly", "t.co", "is.gd",
	"buff.ly", "adf.ly", "shorte.st", "bc.vc", "cutt.ly", "shorturl.at",

	"t.cn", "sinaurl.cn", "dwz.cn", "url.cn", "suo.im", "mrw.so",
	"x.co", "zzb.bz", "985.so", "6du.in", "ft12.com", "url7.me",

	// other short chain
	"rebrand.ly", "bit.do", "short.io", "shrink.me", "short.cm",
}

func isHTMLOK(content string) bool {
	// Checks if the HTML tag is closed
	openTags := regexp.MustCompile(`<([a-zA-Z]+)[^>]*?>`).FindAllStringSubmatch(content, -1)
	closeTags := regexp.MustCompile(`</([a-zA-Z]+)>`).FindAllStringSubmatch(content, -1)
	openCount := make(map[string]int)
	closeCount := make(map[string]int)
	for _, tag := range openTags {
		openCount[strings.ToLower(tag[1])]++
	}
	for _, tag := range closeTags {
		closeCount[strings.ToLower(tag[1])]++
	}
	for tag, count := range openCount {
		if closeCount[tag] != count {
			return false
		}
	}
	return true
}

func isImageOnly(content string) bool {
	imgTag := regexp.MustCompile(`<img[\s\S]*?>`)
	text := regexp.MustCompile(`<.*?>`).ReplaceAllString(content, "")
	return imgTag.MatchString(content) && len(strings.TrimSpace(text)) < 10
}

func hasShortLink(content string) bool {
	for _, domain := range shortLinkDomains {
		if strings.Contains(content, domain) {
			return true
		}
	}
	return false
}
func imgWithoutAltCount(content string) (count int) {
	imgTag := regexp.MustCompile(`<img[^>]*?>`)
	altAttr := regexp.MustCompile(`alt\s*=\s*['\"]`)
	imgs := imgTag.FindAllString(content, -1)
	for _, img := range imgs {
		if !altAttr.MatchString(img) {
			count++
		}
	}
	return
}
func (c *ControllerV1) CheckEmailContent(ctx context.Context, req *v1.CheckEmailContentReq) (res *v1.CheckEmailContentRes, err error) {
	res = &v1.CheckEmailContentRes{}
	var totalScore float64 = 10.0
	var items []ScoreItem
	content := req.Content

	// 1. Spam detection
	for _, word := range spamWords {
		if strings.Contains(strings.ToLower(content), word) {
			items = append(items, ScoreItem{"Spam Word", fmt.Sprintf("Contains spam word: %s", word), -1.0})
			totalScore -= 1.0
		}
	}
	// 2. HTML standard check
	if !isHTMLOK(content) {
		items = append(items, ScoreItem{"HTML Format", "HTML tags are not standard", -0.5})
		totalScore -= 0.5
	}
	if isImageOnly(content) {
		items = append(items, ScoreItem{"Image Ratio", "Content is purely images or image-heavy", -1.0})
		totalScore -= 1.0
	}
	// 2.1 Check for missing alt attributes in img tags
	imgNoAlt := imgWithoutAltCount(content)
	if imgNoAlt > 0 {
		items = append(items, ScoreItem{"Missing Alt Attribute", fmt.Sprintf("%d <img> tags are missing the alt attribute", imgNoAlt), -0.2 * float64(imgNoAlt)})
		totalScore -= 0.2 * float64(imgNoAlt)
	}
	// 3. Link check
	if hasShortLink(content) {
		items = append(items, ScoreItem{"Short Link", "Contains a short link domain", -0.5})
		totalScore -= 0.5
	}

	if totalScore < 0 {
		totalScore = 0
	}

	res.Data.Score = totalScore
	for _, item := range items {
		res.Data.Items = append(res.Data.Items, v1.ScoreItem{
			Name:   item.Name,
			Detail: item.Detail,
			Score:  item.Score,
		})
	}
	res.SetSuccess(public.LangCtx(ctx, "Successful"))
	return res, nil
}
