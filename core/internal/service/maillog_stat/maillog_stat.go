package maillog_stat

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"github.com/gogf/gf/os/gfsnotify"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Define constants
var (
	defaultMaillogPath     = public.AbsPath(filepath.Join(consts.POSTFIX_MAILLOG_PATH, "mail.log"))
	defaultPostfixMainConf = public.AbsPath(consts.POSTFIX_MAIN_CONF)
)

// GetLastMaillogTimeMillis retrieves the last mail log time in milliseconds
func GetLastMaillogTimeMillis() int64 {
	val, err := g.DB().Model("bm_options").Where("name", "last_maillog_time_millis").Value("value")

	if err != nil {
		g.Log().Error(context.Background(), "GetLastMaillogTimeMillis error:", err.Error())
		return 0
	}

	if val == nil {
		return 0
	}

	return val.Int64()
}

// SetLastMaillogTimeMillis sets the last mail log time in milliseconds
func SetLastMaillogTimeMillis(timeMillis int64) {
	_, err := g.DB().Model("bm_options").OnConflict("name").OnDuplicate(g.Map{
		"value": gdb.Raw("excluded.value"),
	}).Save(g.Map{
		"name":  "last_maillog_time_millis",
		"value": timeMillis,
	})

	if err != nil {
		g.Log().Error(context.Background(), "SetLastMaillogTimeMillis error:", err.Error())
		return
	}
}

type MailRecorfContract interface {
	GetPostfixMessageID() string
	GetLogTimeMillis() int64
}

// MailRecord basic mail record structure
type MailRecord struct {
	PostfixMessageID string `json:"postfix_message_id"`
	LogTimeMillis    int64  `json:"log_time_millis"`
}

// GetPostfixMessageID returns the Postfix message ID
func (m *MailRecord) GetPostfixMessageID() string {
	return m.PostfixMessageID
}

// GetLogTimeMillis returns the log time in milliseconds
func (m *MailRecord) GetLogTimeMillis() int64 {
	return m.LogTimeMillis
}

// MailSendRecord mail sending record
type MailSendRecord struct {
	MailRecord
	Recipient    string  `json:"recipient"`
	MailProvider string  `json:"mail_provider"`
	Status       string  `json:"status"`
	Delay        float64 `json:"delay"`
	Delays       string  `json:"delays"`
	Dsn          string  `json:"dsn"`
	Relay        string  `json:"relay"`
	Description  string  `json:"description"`
}

// MailReceiveRecord mail receiving record
type MailReceiveRecord struct {
	MailRecord
	Recipient   string  `json:"recipient"`
	Status      string  `json:"status"`
	Delay       float64 `json:"delay"`
	Delays      string  `json:"delays"`
	Dsn         string  `json:"dsn"`
	Relay       string  `json:"relay"`
	Description string  `json:"description"`
}

// MailMessageID mail MessageID record
type MailMessageID struct {
	MailRecord
	MessageID string `json:"message_id"`
}

// MailSender mail sender record
type MailSender struct {
	MailRecord
	Sender string `json:"sender"`
	Size   int64  `json:"size"`
}

// MailRemoved mail removed record
type MailRemoved struct {
	MailRecord
}

// MailDeferredRecord mail deferred record
type MailDeferredRecord struct {
	MailRecord
	Delay       float64 `json:"delay"`
	Delays      string  `json:"delays"`
	Dsn         string  `json:"dsn"`
	Relay       string  `json:"relay"`
	Description string  `json:"description"`
}

// MaillogStat mail log statistics analyzer
type MaillogStat struct {
	maillogPath         string
	startTime           int64
	endTime             int64
	doSummary           bool
	ignoreRelays        map[string]struct{}
	ignoreMailAddresses map[string]struct{}
	currentYear         int
	delivered           int
	bounced             int
	deferred            int
	deferredTotal       int
	bounceDetails       map[string]map[string]struct{}
	deferralDetails     map[string]map[string]struct{}
	mailHostname        string
	monthMap            map[string]int
	statusPattern       *regexp.Regexp
	recipientPattern    *regexp.Regexp
	delayPattern        *regexp.Regexp
	delaysPattern       *regexp.Regexp
	dsnPattern          *regexp.Regexp
	relayPattern        *regexp.Regexp
	descriptionPattern  *regexp.Regexp
	messageIDPattern    *regexp.Regexp
	mailRemovedPattern  *regexp.Regexp
	mailSenderPattern   *regexp.Regexp
}

// NewMaillogStat creates a new mail log statistics analyzer
func NewMaillogStat(maillogPath string, startTime, endTime int64, doSummary bool) *MaillogStat {
	if maillogPath == "" {
		maillogPath = defaultMaillogPath
	}

	ms := &MaillogStat{
		maillogPath:         maillogPath,
		startTime:           startTime,
		endTime:             endTime,
		doSummary:           doSummary,
		ignoreRelays:        make(map[string]struct{}),
		ignoreMailAddresses: make(map[string]struct{}),
		monthMap: map[string]int{
			"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4,
			"May": 5, "Jun": 6, "Jul": 7, "Aug": 8,
			"Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
		},
		bounceDetails:   make(map[string]map[string]struct{}),
		deferralDetails: make(map[string]map[string]struct{}),
	}

	// Compile regular expressions
	ms.statusPattern = regexp.MustCompile(`status=(\S+) `)
	ms.recipientPattern = regexp.MustCompile(`to=<([^>]+)>`)
	ms.delayPattern = regexp.MustCompile(`delay=(\d+(?:\.\d+)?),`)
	ms.delaysPattern = regexp.MustCompile(`delays=(\d+(?:\.\d+)?(?:/\d+(?:\.\d+)?){3}),`)
	ms.dsnPattern = regexp.MustCompile(`dsn=([^,]+),`)
	ms.relayPattern = regexp.MustCompile(`relay=([^,]+),`)
	ms.descriptionPattern = regexp.MustCompile(`\((.*?)\)$`)
	ms.messageIDPattern = regexp.MustCompile(`postfix/[^\[]+\[\d+]: *([^:]+):`)
	ms.mailRemovedPattern = regexp.MustCompile(`postfix/qmgr\[\d+]: *([^:]+): *removed$`)
	ms.mailSenderPattern = regexp.MustCompile(`postfix/qmgr\[\d+]: *([^:]+): *from=<([^>]+)>, +size=(\d+),`)

	ms.resetData()
	ms.initIgnoreRelays()
	ms.initIgnoreMailAddresses()

	return ms
}

// resetData resets statistical data
func (ms *MaillogStat) resetData() {
	ms.bounceDetails = make(map[string]map[string]struct{})
	ms.deferralDetails = make(map[string]map[string]struct{})

	if fi, err := os.Stat(ms.maillogPath); err == nil {
		ms.currentYear = fi.ModTime().Year()
	}

	ms.delivered = 0
	ms.bounced = 0
	ms.deferred = 0
	ms.deferredTotal = 0
}

// initIgnoreRelays initializes ignored relays
func (ms *MaillogStat) initIgnoreRelays() {
	if _, err := os.Stat(defaultPostfixMainConf); err != nil {
		return
	}

	content, err := os.ReadFile(defaultPostfixMainConf)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	var myhostname, virtualTransport string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if myhostname != "" && virtualTransport != "" {
			break
		}

		if myhostname == "" {
			if matches := regexp.MustCompile(`myhostname *= *([^\r\n]+)`).FindStringSubmatch(line); matches != nil {
				myhostname = matches[1]
				continue
			}
		}

		if virtualTransport == "" {
			if matches := regexp.MustCompile(`virtual_transport *= *([^\r\n]+)`).FindStringSubmatch(line); matches != nil {
				virtualTransport = matches[1]
				continue
			}
		}
	}

	if myhostname != "" && virtualTransport != "" {
		ms.mailHostname = myhostname
		relay := fmt.Sprintf("%s[%s]", myhostname, strings.Split(virtualTransport, ":")[len(strings.Split(virtualTransport, ":"))-1])
		ms.ignoreRelays[relay] = struct{}{}
	}
}

// initIgnoreMailAddresses initializes ignored mail addresses
func (ms *MaillogStat) initIgnoreMailAddresses() {
	ms.ignoreMailAddresses["root@localhost"] = struct{}{}
	if ms.mailHostname != "" {
		ms.ignoreMailAddresses[fmt.Sprintf("root@%s", ms.mailHostname)] = struct{}{}
	}
}

// getMailProvider gets mail service provider
func (ms *MaillogStat) getMailProvider(email string) string {
	return public.GetMailProviderGroup(email)
}

// parseLogTimeMillis parses log time (millisecond timestamp)
func (ms *MaillogStat) parseLogTimeMillis(line string) int64 {
	line = strings.TrimSpace(line)

	// Try to parse ISO 8601 format
	if len(line) >= 4 && isDigit(line[:4]) {
		t, err := time.Parse(time.RFC3339, strings.Split(line, " ")[0])
		if err == nil {
			return t.UnixMilli()
		}
	}

	parts := strings.Fields(line)
	if len(parts) < 4 {
		return -1
	}

	month, ok := ms.monthMap[parts[0]]
	if !ok {
		return -1
	}

	day, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1
	}

	timeParts := strings.Split(parts[2], ":")
	if len(timeParts) != 3 {
		return -1
	}

	hour, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return -1
	}

	minute, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return -1
	}

	second, err := strconv.Atoi(timeParts[2])
	if err != nil {
		return -1
	}

	t := time.Date(ms.currentYear, time.Month(month), day, hour, minute, second, 0, time.Local)
	return t.UnixMilli()
}

// isDigit checks if string consists of digits only
func isDigit(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Analysis analyzes mail logs
func (ms *MaillogStat) Analysis(ctx context.Context) (<-chan MailRecorfContract, error) {
	ms.resetData()

	if _, err := os.Stat(ms.maillogPath); err != nil {
		return nil, err
	}

	recordChan := make(chan MailRecorfContract, 1000)
	go func() {
		defer close(recordChan)

		n := 0

		err := public.ReadEachReverse(ms.maillogPath, func(line string, cnt int) bool {
			n++

			select {
			case <-ctx.Done():
				return false
			default:
				record, stop := ms.analyzeLine(line)

				if stop {
					return false
				}

				if record == nil || record.GetLogTimeMillis() < 1 {
					return true
				}

				if r, ok := record.(*MailSendRecord); ok && r.Status == "deferred" {
					recordChan <- &MailDeferredRecord{
						MailRecord: MailRecord{
							PostfixMessageID: r.PostfixMessageID,
							LogTimeMillis:    r.LogTimeMillis,
						},
						Delay:       r.Delay,
						Delays:      r.Delays,
						Dsn:         r.Dsn,
						Relay:       r.Relay,
						Description: r.Description,
					}
				}

				recordChan <- record
			}
			return true
		})

		g.Log().Debug(ctx, "Total scanned lines:", n)

		if err != nil {
			g.Log().Warning(ctx, "Error reading maillog: %v\n", err)
		}
	}()

	return recordChan, nil
}

// analyzeLine analyzes single log line
func (ms *MaillogStat) analyzeLine(line string) (record MailRecorfContract, stop bool) {
	line = strings.TrimSpace(line)
	logTime := ms.parseLogTimeMillis(line)

	if logTime < 1 {
		return
	}

	if ms.endTime > 0 && logTime > ms.endTime {
		return
	}

	if ms.startTime > 0 && logTime < ms.startTime {
		stop = true
		return
	}

	// Analyze different types of logs
	if strings.Contains(line, "postfix/lmtp[") {
		record = ms.analyzeReceiveMail(line, logTime)
		return
	}

	if strings.Contains(line, "postfix/qmgr[") {
		if strings.Contains(line, "from=<") {
			record = ms.analyzeSender(line, logTime)
			return
		}

		if strings.Contains(line, "removed") {
			record = ms.analyzeRemoved(line, logTime)
			return
		}
	}

	if strings.Contains(line, "postfix/cleanup[") && strings.Contains(line, "message-id=<") {
		record = ms.analyzeMessageID(line, logTime)
		return
	}

	if strings.Contains(line, "to=<") && strings.Contains(line, "status=") && strings.Contains(line, "dsn=") {
		record = ms.analyzeSendMail(line, logTime)
		return
	}

	record = &MailRecord{
		PostfixMessageID: "",
		LogTimeMillis:    logTime,
	}

	return
}

// analyzeReceiveMail analyzes mail receiving logs
func (ms *MaillogStat) analyzeReceiveMail(line string, logTime int64) MailRecorfContract {
	matches := ms.messageIDPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	recipientMatches := ms.recipientPattern.FindStringSubmatch(line)
	if recipientMatches == nil {
		return nil
	}

	statusMatches := ms.statusPattern.FindStringSubmatch(line)
	if statusMatches == nil {
		return nil
	}

	record := &MailReceiveRecord{
		MailRecord: MailRecord{
			PostfixMessageID: matches[1],
			LogTimeMillis:    logTime,
		},
		Recipient: recipientMatches[1],
		Status:    statusMatches[1],
	}

	if _, ok := ms.ignoreMailAddresses[record.Recipient]; ok {
		return nil
	}

	if relayMatches := ms.relayPattern.FindStringSubmatch(line); relayMatches != nil {
		record.Relay = relayMatches[1]
	}

	if delayMatches := ms.delayPattern.FindStringSubmatch(line); delayMatches != nil {
		record.Delay, _ = strconv.ParseFloat(delayMatches[1], 64)
	}

	if delaysMatches := ms.delaysPattern.FindStringSubmatch(line); delaysMatches != nil {
		record.Delays = delaysMatches[1]
	}

	if dsnMatches := ms.dsnPattern.FindStringSubmatch(line); dsnMatches != nil {
		record.Dsn = dsnMatches[1]
	}

	if descMatches := ms.descriptionPattern.FindStringSubmatch(line); descMatches != nil {
		record.Description = descMatches[1]
	}

	return record
}

// analyzeSendMail analyzes mail sending logs
func (ms *MaillogStat) analyzeSendMail(line string, logTime int64) MailRecorfContract {
	matches := ms.messageIDPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	recipientMatches := ms.recipientPattern.FindStringSubmatch(line)
	if recipientMatches == nil {
		return nil
	}

	statusMatches := ms.statusPattern.FindStringSubmatch(line)
	if statusMatches == nil {
		return nil
	}

	record := &MailSendRecord{
		MailRecord: MailRecord{
			PostfixMessageID: matches[1],
			LogTimeMillis:    logTime,
		},
		Recipient: recipientMatches[1],
		Status:    statusMatches[1],
	}

	if _, ok := ms.ignoreMailAddresses[record.Recipient]; ok {
		return nil
	}

	if relayMatches := ms.relayPattern.FindStringSubmatch(line); relayMatches != nil {
		relay := relayMatches[1]
		if _, ok := ms.ignoreRelays[relay]; ok {
			return nil
		}
		record.Relay = relay
	}

	if delayMatches := ms.delayPattern.FindStringSubmatch(line); delayMatches != nil {
		record.Delay, _ = strconv.ParseFloat(delayMatches[1], 64)
	}

	if delaysMatches := ms.delaysPattern.FindStringSubmatch(line); delaysMatches != nil {
		record.Delays = delaysMatches[1]
	}

	if dsnMatches := ms.dsnPattern.FindStringSubmatch(line); dsnMatches != nil {
		record.Dsn = dsnMatches[1]
	}

	if descMatches := ms.descriptionPattern.FindStringSubmatch(line); descMatches != nil {
		record.Description = descMatches[1]
	}

	record.MailProvider = ms.getMailProvider(record.Recipient)

	switch record.Status {
	case "sent":
		ms.delivered++
	case "bounced":
		ms.bounced++
		if _, ok := ms.bounceDetails[record.Dsn]; !ok {
			ms.bounceDetails[record.Dsn] = make(map[string]struct{})
		}
		ms.bounceDetails[record.Dsn][record.Description] = struct{}{}
	case "deferred":
		ms.deferred++
		ms.deferredTotal++
		if _, ok := ms.deferralDetails[record.Dsn]; !ok {
			ms.deferralDetails[record.Dsn] = make(map[string]struct{})
		}
		ms.deferralDetails[record.Dsn][record.Description] = struct{}{}
	}

	return record
}

// analyzeMessageID analyzes MessageID logs
func (ms *MaillogStat) analyzeMessageID(line string, logTime int64) MailRecorfContract {
	matches := ms.messageIDPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	messageIDMatches := regexp.MustCompile(`message-id=<([^>]+)>`).FindStringSubmatch(line)
	if messageIDMatches == nil {
		return nil
	}

	return &MailMessageID{
		MailRecord: MailRecord{
			PostfixMessageID: matches[1],
			LogTimeMillis:    logTime,
		},
		MessageID: messageIDMatches[1],
	}
}

// analyzeRemoved analyzes mail removed logs
func (ms *MaillogStat) analyzeRemoved(line string, logTime int64) MailRecorfContract {
	matches := ms.mailRemovedPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	return &MailRemoved{
		MailRecord: MailRecord{
			PostfixMessageID: matches[1],
			LogTimeMillis:    logTime,
		},
	}
}

// analyzeSender analyzes sender logs
func (ms *MaillogStat) analyzeSender(line string, logTime int64) MailRecorfContract {
	matches := ms.mailSenderPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	size, err := strconv.ParseInt(matches[3], 10, 64)
	if err != nil {
		return nil
	}

	return &MailSender{
		MailRecord: MailRecord{
			PostfixMessageID: matches[1],
			LogTimeMillis:    logTime,
		},
		Sender: matches[2],
		Size:   size,
	}
}

// AnalysisAndSaveToDatabase analyzes mail logs and saves to database
func (ms *MaillogStat) AnalysisAndSaveToDatabase(ctx context.Context) error {
	recordChan, err := ms.Analysis(ctx)
	if err != nil {
		return err
	}

	lastMaillogTime := GetLastMaillogTimeMillis()
	sendRecords := make([]*MailSendRecord, 0, 256)
	receiveRecords := make([]*MailReceiveRecord, 0, 256)
	messageIDRecords := make([]*MailMessageID, 0, 256)
	senderRecords := make([]*MailSender, 0, 256)
	removedRecords := make([]*MailRemoved, 0, 256)
	deferredRecords := make([]*MailDeferredRecord, 0, 256)

	existsPostfixMessageIdsForSends := make(map[string]struct{}, 256)
	existsPostfixMessageIdsForReceives := make(map[string]struct{}, 256)

	for record := range recordChan {
		// g.Log().Debug(context.Background(), "AnalysisAndSaveToDatabase: record", record)

		if record.GetLogTimeMillis() > lastMaillogTime {
			lastMaillogTime = record.GetLogTimeMillis()
		}

		if record.GetPostfixMessageID() == "" {
			continue
		}

		switch r := record.(type) {
		case *MailSendRecord:
			if _, ok := existsPostfixMessageIdsForSends[r.GetPostfixMessageID()]; ok {
				continue
			}
			r.Description = public.SanitizeUTF8(r.Description)
			r.Relay = public.SanitizeUTF8(r.Relay)
			r.Dsn = public.SanitizeUTF8(r.Dsn)
			sendRecords = append(sendRecords, r)
			existsPostfixMessageIdsForSends[r.GetPostfixMessageID()] = struct{}{}
		case *MailReceiveRecord:
			if _, ok := existsPostfixMessageIdsForReceives[r.GetPostfixMessageID()]; ok {
				continue
			}
			r.Description = public.SanitizeUTF8(r.Description)
			r.Relay = public.SanitizeUTF8(r.Relay)
			r.Dsn = public.SanitizeUTF8(r.Dsn)
			receiveRecords = append(receiveRecords, r)
			existsPostfixMessageIdsForReceives[r.GetPostfixMessageID()] = struct{}{}
		case *MailSender:
			senderRecords = append(senderRecords, r)
		case *MailRemoved:
			removedRecords = append(removedRecords, r)
		case *MailMessageID:
			messageIDRecords = append(messageIDRecords, r)
		case *MailDeferredRecord:
			r.Description = public.SanitizeUTF8(r.Description)
			r.Relay = public.SanitizeUTF8(r.Relay)
			r.Dsn = public.SanitizeUTF8(r.Dsn)
			deferredRecords = append(deferredRecords, r)
		default:
			continue
		}
	}

	if lastMaillogTime > 0 {
		g.Log().Debug(context.Background(), "AnalysisAndSaveToDatabase: lastMaillogTime", lastMaillogTime)
		SetLastMaillogTimeMillis(lastMaillogTime)
	}

	if len(sendRecords) > 0 {
		g.Log().Debug(context.Background(), "AnalysisAndSaveToDatabase: sendRecords", len(sendRecords))
		_, err = g.DB().Model("mailstat_send_mails").OnConflict("postfix_message_id").OnDuplicate(g.Map{
			"status":          gdb.Raw("case when excluded.log_time_millis > mailstat_send_mails.log_time_millis then excluded.status else mailstat_send_mails.status end"),
			"delay":           gdb.Raw("case when excluded.log_time_millis > mailstat_send_mails.log_time_millis then excluded.delay else mailstat_send_mails.delay end"),
			"delays":          gdb.Raw("case when excluded.log_time_millis > mailstat_send_mails.log_time_millis then excluded.delays else mailstat_send_mails.delays end"),
			"dsn":             gdb.Raw("case when excluded.log_time_millis > mailstat_send_mails.log_time_millis then excluded.dsn else mailstat_send_mails.dsn end"),
			"relay":           gdb.Raw("case when excluded.log_time_millis > mailstat_send_mails.log_time_millis then excluded.relay else mailstat_send_mails.relay end"),
			"description":     gdb.Raw("case when excluded.log_time_millis > mailstat_send_mails.log_time_millis then excluded.description else mailstat_send_mails.description end"),
			"log_time_millis": gdb.Raw("case when excluded.log_time_millis > mailstat_send_mails.log_time_millis then excluded.log_time_millis else mailstat_send_mails.log_time_millis end"),
		}).Batch(5000).Save(sendRecords)
		if err != nil {
			g.Log().Error(context.Background(), err.Error())
		}
	}

	if len(receiveRecords) > 0 {
		g.Log().Debug(context.Background(), "AnalysisAndSaveToDatabase: receiveRecords", len(receiveRecords))
		_, err = g.DB().Model("mailstat_receive_mails").OnConflict("postfix_message_id").OnDuplicate(g.Map{
			"status":          gdb.Raw("case when excluded.log_time_millis > mailstat_receive_mails.log_time_millis then excluded.status else mailstat_receive_mails.status end"),
			"delay":           gdb.Raw("case when excluded.log_time_millis > mailstat_receive_mails.log_time_millis then excluded.delay else mailstat_receive_mails.delay end"),
			"delays":          gdb.Raw("case when excluded.log_time_millis > mailstat_receive_mails.log_time_millis then excluded.delays else mailstat_receive_mails.delays end"),
			"dsn":             gdb.Raw("case when excluded.log_time_millis > mailstat_receive_mails.log_time_millis then excluded.dsn else mailstat_receive_mails.dsn end"),
			"relay":           gdb.Raw("case when excluded.log_time_millis > mailstat_receive_mails.log_time_millis then excluded.relay else mailstat_receive_mails.relay end"),
			"description":     gdb.Raw("case when excluded.log_time_millis > mailstat_receive_mails.log_time_millis then excluded.description else mailstat_receive_mails.description end"),
			"log_time_millis": gdb.Raw("case when excluded.log_time_millis > mailstat_receive_mails.log_time_millis then excluded.log_time_millis else mailstat_receive_mails.log_time_millis end"),
		}).Batch(5000).Save(receiveRecords)
		if err != nil {
			g.Log().Error(context.Background(), err.Error())
		}
	}

	if len(senderRecords) > 0 {
		g.Log().Debug(context.Background(), "AnalysisAndSaveToDatabase: sendRecords", len(senderRecords))
		_, err = g.DB().Model("mailstat_senders").Batch(5000).InsertIgnore(senderRecords)
		if err != nil {
			g.Log().Error(context.Background(), err.Error())
		}
	}

	if len(removedRecords) > 0 {
		g.Log().Debug(context.Background(), "AnalysisAndSaveToDatabase: removedRecords", len(removedRecords))
		_, err = g.DB().Model("mailstat_removed").Batch(5000).InsertIgnore(removedRecords)
		if err != nil {
			g.Log().Error(context.Background(), err.Error())
		}
	}

	if len(messageIDRecords) > 0 {
		g.Log().Debug(context.Background(), "AnalysisAndSaveToDatabase: messageIDRecords", len(messageIDRecords))
		_, err = g.DB().Model("mailstat_message_ids").Batch(5000).InsertIgnore(messageIDRecords)
		if err != nil {
			g.Log().Error(context.Background(), err.Error())
		}
	}

	if len(deferredRecords) > 0 {
		g.Log().Debug(context.Background(), "AnalysisAndSaveToDatabase: deferredRecords", len(deferredRecords))
		_, err = g.DB().Model("mailstat_deferred_mails").Batch(5000).Insert(deferredRecords)
		if err != nil {
			g.Log().Error(context.Background(), err.Error())
		}
	}

	return nil
}

// MallogEventHandler handles mail log events
type MallogEventHandler struct {
	maillogStat *MaillogStat
	delay       time.Duration
	timer       *time.Timer
	isHanding   bool
}

// NewMallogEventHandler creates a new mail log event handler
func NewMallogEventHandler(maillogPath string, delay time.Duration) *MallogEventHandler {
	return &MallogEventHandler{
		maillogStat: NewMaillogStat(maillogPath, GetLastMaillogTimeMillis(), 0, false),
		delay:       delay,
		timer:       time.NewTimer(delay),
		isHanding:   false,
	}
}

// Start starts the mail log event handler
func (handler *MallogEventHandler) Start() {
	_, err := gfsnotify.Add(handler.maillogStat.maillogPath, func(event *gfsnotify.Event) {
		// g.Log().Debug(context.Background(), "MallogEventHandler: catch event", event)

		if event.IsWrite() {
			// g.Log().Debug(context.Background(), "MallogEventHandler: write event", event)
			if event.Path == handler.maillogStat.maillogPath {
				// g.Log().Debug(context.Background(), "Reset timer", event)
				//if !handler.timer.Stop() {
				//	<-handler.timer.C
				//}
				if !handler.isHanding {
					handler.isHanding = true
					handler.timer.Reset(handler.delay)
				}
			}
		}

		//if event.IsRename() {
		//	g.Log().Debug(context.Background(), "MallogEventHandler: rename event", event)
		//}
	})

	if err != nil {
		g.Log().Error(context.Background(), "MallogEventHandler: error adding file watcher", err)
		return
	}

	g.Log().Debug(context.Background(), "MallogEventHandler: start event")

	for {
		select {
		case <-handler.timer.C:
			handler.isHanding = false
			g.Log().Debug(context.Background(), "MallogEventHandler: timer event")
			handler.maillogStat.startTime = GetLastMaillogTimeMillis()
			if err = handler.maillogStat.AnalysisAndSaveToDatabase(context.Background()); err != nil {
				g.Log().Error(context.Background(), "MallogEventHandler: error analysing maillog", err)
			}
		}
	}
}

// SearchPostfixMessageIdByMessageId searches for Postfix message ID by Message ID
func SearchPostfixMessageIdByMessageId(messageId string) (string, error) {
	val, err := g.DB().Model("mailstat_message_ids").Where("message_id", messageId).Value("postfix_message_id")
	if err != nil {
		return "", err
	}

	return val.String(), nil
}
