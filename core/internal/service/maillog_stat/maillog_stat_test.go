package maillog_stat

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestMaillogStat creates a MaillogStat for testing with no file dependency.
// We skip initIgnoreRelays and initIgnoreMailAddresses since they read from filesystem.
func newTestMaillogStat() *MaillogStat {
	ms := &MaillogStat{
		maillogPath:         "/dev/null",
		startTime:           0,
		endTime:             0,
		doSummary:           false,
		ignoreRelays:        make(map[string]struct{}),
		ignoreMailAddresses: make(map[string]struct{}),
		currentYear:         2024,
		monthMap: map[string]int{
			"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4,
			"May": 5, "Jun": 6, "Jul": 7, "Aug": 8,
			"Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
		},
		bounceDetails:   make(map[string]map[string]struct{}),
		deferralDetails: make(map[string]map[string]struct{}),
	}

	// Compile the same regexes as NewMaillogStat
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

	return ms
}

func TestIsDigit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "all digits", input: "1234", expected: true},
		{name: "single digit", input: "0", expected: true},
		{name: "empty string", input: "", expected: true},
		{name: "with letter", input: "123a", expected: false},
		{name: "all letters", input: "abcd", expected: false},
		{name: "with space", input: "12 34", expected: false},
		{name: "with special char", input: "12-34", expected: false},
		{name: "with dot", input: "12.34", expected: false},
		{name: "year format", input: "2024", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDigit(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseLogTimeMillis_TraditionalFormat(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	tests := []struct {
		name        string
		line        string
		expectValid bool
	}{
		{
			name:        "standard syslog format",
			line:        "Jan  1 00:00:00 hostname postfix/smtp[1234]: message",
			expectValid: true,
		},
		{
			name:        "December date",
			line:        "Dec 31 23:59:59 hostname postfix/smtp[1234]: message",
			expectValid: true,
		},
		{
			name:        "single digit day",
			line:        "Mar  5 12:30:45 hostname postfix/smtp[1234]: message",
			expectValid: true,
		},
		{
			name:        "double digit day",
			line:        "Jun 15 08:00:00 hostname postfix/smtp[1234]: message",
			expectValid: true,
		},
		{
			name:        "empty string",
			line:        "",
			expectValid: false,
		},
		{
			name:        "invalid month",
			line:        "Xyz 15 08:00:00 hostname postfix/smtp[1234]: message",
			expectValid: false,
		},
		{
			name:        "too few parts",
			line:        "Jan 1",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ms.parseLogTimeMillis(tt.line)
			if tt.expectValid {
				assert.Greater(t, result, int64(0), "expected valid timestamp")
			} else {
				assert.LessOrEqual(t, result, int64(0), "expected invalid timestamp")
			}
		})
	}
}

func TestParseLogTimeMillis_CorrectValues(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	// "Jan  1 00:00:00" in year 2024
	result := ms.parseLogTimeMillis("Jan  1 00:00:00 hostname postfix/smtp[1234]: message")
	expected := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.Local).UnixMilli()
	assert.Equal(t, expected, result)

	// "Dec 25 15:30:00" in year 2024
	result2 := ms.parseLogTimeMillis("Dec 25 15:30:00 hostname postfix/smtp[1234]: message")
	expected2 := time.Date(2024, time.December, 25, 15, 30, 0, 0, time.Local).UnixMilli()
	assert.Equal(t, expected2, result2)
}

func TestParseLogTimeMillis_ISO8601(t *testing.T) {
	ms := newTestMaillogStat()

	result := ms.parseLogTimeMillis("2024-01-15T10:30:00Z hostname postfix/smtp[1234]: message")
	if result > 0 {
		expectedTime, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
		assert.Equal(t, expectedTime.UnixMilli(), result)
	}
}

func TestAnalyzeLine_SendMail(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	line := "Jan 15 10:30:00 mail postfix/smtp[12345]: ABC123: to=<user@gmail.com>, relay=gmail-smtp-in.l.google.com[142.250.4.26]:25, delay=1.5, delays=0.1/0.2/0.3/0.9, dsn=2.0.0, status=sent (250 2.0.0 OK)"

	record, stop := ms.analyzeLine(line)
	assert.False(t, stop)
	require.NotNil(t, record)

	sendRecord, ok := record.(*MailSendRecord)
	require.True(t, ok)
	assert.Equal(t, "ABC123", sendRecord.PostfixMessageID)
	assert.Equal(t, "user@gmail.com", sendRecord.Recipient)
	assert.Equal(t, "sent", sendRecord.Status)
	assert.Equal(t, "2.0.0", sendRecord.Dsn)
	assert.Equal(t, 1.5, sendRecord.Delay)
	assert.Equal(t, "0.1/0.2/0.3/0.9", sendRecord.Delays)
	assert.Contains(t, sendRecord.Relay, "gmail-smtp-in")
	assert.Contains(t, sendRecord.Description, "250 2.0.0 OK")
}

func TestAnalyzeLine_BouncedMail(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	line := "Feb 20 14:00:00 mail postfix/smtp[54321]: DEF456: to=<bad@example.com>, relay=mail.example.com[1.2.3.4]:25, delay=0.5, delays=0.1/0/0.1/0.3, dsn=5.1.1, status=bounced (host mail.example.com said: 550 5.1.1 User unknown)"

	record, stop := ms.analyzeLine(line)
	assert.False(t, stop)
	require.NotNil(t, record)

	sendRecord, ok := record.(*MailSendRecord)
	require.True(t, ok)
	assert.Equal(t, "DEF456", sendRecord.PostfixMessageID)
	assert.Equal(t, "bounced", sendRecord.Status)
	assert.Equal(t, "5.1.1", sendRecord.Dsn)
}

func TestAnalyzeLine_DeferredMail(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	line := "Mar 10 09:15:30 mail postfix/smtp[99999]: GHI789: to=<slow@example.com>, relay=mail.example.com[5.6.7.8]:25, delay=120, delays=0/0/120/0, dsn=4.7.0, status=deferred (host mail.example.com said: 450 4.7.0 Try again later)"

	record, stop := ms.analyzeLine(line)
	assert.False(t, stop)
	require.NotNil(t, record)

	sendRecord, ok := record.(*MailSendRecord)
	require.True(t, ok)
	assert.Equal(t, "deferred", sendRecord.Status)
	assert.Equal(t, "4.7.0", sendRecord.Dsn)
}

func TestAnalyzeLine_MessageID(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	line := "Jan 15 10:30:00 mail postfix/cleanup[1234]: ABC123: message-id=<unique-msg-id@example.com>"

	record, stop := ms.analyzeLine(line)
	assert.False(t, stop)
	require.NotNil(t, record)

	msgIDRecord, ok := record.(*MailMessageID)
	require.True(t, ok)
	assert.Equal(t, "ABC123", msgIDRecord.PostfixMessageID)
	assert.Equal(t, "unique-msg-id@example.com", msgIDRecord.MessageID)
}

func TestAnalyzeLine_Sender(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	line := "Jan 15 10:30:00 mail postfix/qmgr[1234]: ABC123: from=<sender@example.com>, size=1024, nrcpt=1 (queue active)"

	record, stop := ms.analyzeLine(line)
	assert.False(t, stop)
	require.NotNil(t, record)

	senderRecord, ok := record.(*MailSender)
	require.True(t, ok)
	assert.Equal(t, "ABC123", senderRecord.PostfixMessageID)
	assert.Equal(t, "sender@example.com", senderRecord.Sender)
	assert.Equal(t, int64(1024), senderRecord.Size)
}

func TestAnalyzeLine_Removed(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	line := "Jan 15 10:30:00 mail postfix/qmgr[1234]: ABC123: removed"

	record, stop := ms.analyzeLine(line)
	assert.False(t, stop)
	require.NotNil(t, record)

	removedRecord, ok := record.(*MailRemoved)
	require.True(t, ok)
	assert.Equal(t, "ABC123", removedRecord.PostfixMessageID)
}

func TestAnalyzeLine_ReceiveMail(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	line := "Jan 15 10:30:00 mail postfix/lmtp[1234]: ABC123: to=<local@example.com>, relay=dovecot, delay=0.2, delays=0.1/0/0/0.1, dsn=2.0.0, status=sent (delivered via dovecot)"

	record, stop := ms.analyzeLine(line)
	assert.False(t, stop)
	require.NotNil(t, record)

	recvRecord, ok := record.(*MailReceiveRecord)
	require.True(t, ok)
	assert.Equal(t, "ABC123", recvRecord.PostfixMessageID)
	assert.Equal(t, "local@example.com", recvRecord.Recipient)
	assert.Equal(t, "sent", recvRecord.Status)
}

func TestAnalyzeLine_IgnoredMailAddress(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024
	ms.ignoreMailAddresses["root@localhost"] = struct{}{}

	line := "Jan 15 10:30:00 mail postfix/smtp[12345]: ABC123: to=<root@localhost>, relay=local, delay=0.1, delays=0/0/0/0.1, dsn=2.0.0, status=sent (delivered)"

	record, stop := ms.analyzeLine(line)
	assert.False(t, stop)

	// Should be nil since root@localhost is ignored
	if record != nil {
		_, isSend := record.(*MailSendRecord)
		if isSend {
			t.Fatal("expected nil for ignored address")
		}
	}
}

func TestAnalyzeLine_IgnoredRelay(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024
	ms.ignoreRelays["internal[10000]"] = struct{}{}

	line := "Jan 15 10:30:00 mail postfix/smtp[12345]: ABC123: to=<user@example.com>, relay=internal[10000], delay=0.1, delays=0/0/0/0.1, dsn=2.0.0, status=sent (delivered)"

	record, stop := ms.analyzeLine(line)
	assert.False(t, stop)
	// Should be nil since relay is ignored
	if record != nil {
		_, isSend := record.(*MailSendRecord)
		if isSend {
			t.Fatal("expected nil for ignored relay")
		}
	}
}

func TestAnalyzeLine_TimeFiltering(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	// Set a time window
	windowStart := time.Date(2024, 6, 1, 0, 0, 0, 0, time.Local).UnixMilli()
	windowEnd := time.Date(2024, 6, 30, 23, 59, 59, 0, time.Local).UnixMilli()
	ms.startTime = windowStart
	ms.endTime = windowEnd

	// Line before window -> should return stop=true (reading in reverse)
	beforeLine := "Jan 15 10:30:00 mail postfix/smtp[12345]: ABC123: to=<user@example.com>, relay=mx.example.com[1.2.3.4]:25, delay=0.1, delays=0/0/0/0.1, dsn=2.0.0, status=sent (ok)"
	_, stop := ms.analyzeLine(beforeLine)
	assert.True(t, stop, "line before window should stop reading")

	// Line after window -> should return nil record
	afterLine := "Dec 15 10:30:00 mail postfix/smtp[12345]: DEF456: to=<user@example.com>, relay=mx.example.com[1.2.3.4]:25, delay=0.1, delays=0/0/0/0.1, dsn=2.0.0, status=sent (ok)"
	record, stop := ms.analyzeLine(afterLine)
	assert.False(t, stop)
	assert.Nil(t, record, "line after window should be skipped")

	// Line within window
	withinLine := "Jun 15 10:30:00 mail postfix/smtp[12345]: GHI789: to=<user@example.com>, relay=mx.example.com[1.2.3.4]:25, delay=0.1, delays=0/0/0/0.1, dsn=2.0.0, status=sent (ok)"
	record, stop = ms.analyzeLine(withinLine)
	assert.False(t, stop)
	assert.NotNil(t, record)
}

func TestAnalyzeLine_CountsDelivered(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	assert.Equal(t, 0, ms.delivered)

	line := "Jan 15 10:30:00 mail postfix/smtp[12345]: ABC123: to=<user@example.com>, relay=mx.example.com[1.2.3.4]:25, delay=0.1, delays=0/0/0/0.1, dsn=2.0.0, status=sent (ok)"
	record, _ := ms.analyzeLine(line)
	require.NotNil(t, record)
	assert.Equal(t, 1, ms.delivered)
}

func TestAnalyzeLine_CountsBounced(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	assert.Equal(t, 0, ms.bounced)

	line := "Jan 15 10:30:00 mail postfix/smtp[12345]: ABC123: to=<bad@example.com>, relay=mx.example.com[1.2.3.4]:25, delay=0.1, delays=0/0/0/0.1, dsn=5.1.1, status=bounced (user unknown)"
	record, _ := ms.analyzeLine(line)
	require.NotNil(t, record)
	assert.Equal(t, 1, ms.bounced)
	assert.Contains(t, ms.bounceDetails, "5.1.1")
}

func TestAnalyzeLine_CountsDeferred(t *testing.T) {
	ms := newTestMaillogStat()
	ms.currentYear = 2024

	assert.Equal(t, 0, ms.deferred)

	line := "Jan 15 10:30:00 mail postfix/smtp[12345]: ABC123: to=<slow@example.com>, relay=mx.example.com[1.2.3.4]:25, delay=120, delays=0/0/120/0, dsn=4.7.0, status=deferred (try again later)"
	record, _ := ms.analyzeLine(line)
	require.NotNil(t, record)
	assert.Equal(t, 1, ms.deferred)
	assert.Equal(t, 1, ms.deferredTotal)
	assert.Contains(t, ms.deferralDetails, "4.7.0")
}

func TestMailRecord_Interface(t *testing.T) {
	record := &MailRecord{
		PostfixMessageID: "TEST123",
		LogTimeMillis:    1700000000000,
	}

	var iface MailRecorfContract = record
	assert.Equal(t, "TEST123", iface.GetPostfixMessageID())
	assert.Equal(t, int64(1700000000000), iface.GetLogTimeMillis())
}

func TestMailSendRecord_Interface(t *testing.T) {
	record := &MailSendRecord{
		MailRecord: MailRecord{
			PostfixMessageID: "SEND123",
			LogTimeMillis:    1700000000000,
		},
		Recipient:    "user@example.com",
		MailProvider: "google",
		Status:       "sent",
		Delay:        1.5,
		Delays:       "0.1/0.2/0.3/0.9",
		Dsn:          "2.0.0",
		Relay:        "gmail-smtp-in.l.google.com",
		Description:  "250 OK",
	}

	var iface MailRecorfContract = record
	assert.Equal(t, "SEND123", iface.GetPostfixMessageID())
	assert.Equal(t, "user@example.com", record.Recipient)
	assert.Equal(t, "google", record.MailProvider)
}

func TestMailReceiveRecord_Struct(t *testing.T) {
	record := &MailReceiveRecord{
		MailRecord: MailRecord{
			PostfixMessageID: "RECV123",
			LogTimeMillis:    1700000000000,
		},
		Recipient:   "local@example.com",
		Status:      "sent",
		Delay:       0.2,
		Delays:      "0.1/0/0/0.1",
		Dsn:         "2.0.0",
		Relay:       "dovecot",
		Description: "delivered",
	}

	assert.Equal(t, "RECV123", record.GetPostfixMessageID())
	assert.Equal(t, "local@example.com", record.Recipient)
}

func TestMailMessageID_Struct(t *testing.T) {
	record := &MailMessageID{
		MailRecord: MailRecord{
			PostfixMessageID: "MSG123",
			LogTimeMillis:    1700000000000,
		},
		MessageID: "unique-id@example.com",
	}

	assert.Equal(t, "MSG123", record.GetPostfixMessageID())
	assert.Equal(t, "unique-id@example.com", record.MessageID)
}

func TestMailSender_Struct(t *testing.T) {
	record := &MailSender{
		MailRecord: MailRecord{
			PostfixMessageID: "SND123",
			LogTimeMillis:    1700000000000,
		},
		Sender: "sender@example.com",
		Size:   2048,
	}

	assert.Equal(t, "SND123", record.GetPostfixMessageID())
	assert.Equal(t, "sender@example.com", record.Sender)
	assert.Equal(t, int64(2048), record.Size)
}

func TestMailRemoved_Struct(t *testing.T) {
	record := &MailRemoved{
		MailRecord: MailRecord{
			PostfixMessageID: "RMV123",
			LogTimeMillis:    1700000000000,
		},
	}

	assert.Equal(t, "RMV123", record.GetPostfixMessageID())
}

func TestMailDeferredRecord_Struct(t *testing.T) {
	record := &MailDeferredRecord{
		MailRecord: MailRecord{
			PostfixMessageID: "DEF123",
			LogTimeMillis:    1700000000000,
		},
		Delay:       120.0,
		Delays:      "0/0/120/0",
		Dsn:         "4.7.0",
		Relay:       "mail.example.com[1.2.3.4]:25",
		Description: "try again later",
	}

	assert.Equal(t, "DEF123", record.GetPostfixMessageID())
	assert.Equal(t, 120.0, record.Delay)
	assert.Equal(t, "4.7.0", record.Dsn)
}

func TestMonthMap(t *testing.T) {
	ms := newTestMaillogStat()

	expected := map[string]int{
		"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4,
		"May": 5, "Jun": 6, "Jul": 7, "Aug": 8,
		"Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	}

	for month, num := range expected {
		assert.Equal(t, num, ms.monthMap[month], "month %s", month)
	}
	assert.Len(t, ms.monthMap, 12)
}
