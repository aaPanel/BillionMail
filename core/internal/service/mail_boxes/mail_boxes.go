package mail_boxes

import (
	v1 "billionmail-core/api/mail_boxes/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/public"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/GehirnInc/crypt/md5_crypt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func Add(ctx context.Context, mailbox *v1.Mailbox) (err error) {
	// Encode password
	mailbox.PasswordEncode = PasswdEncode(ctx, mailbox.Password)

	// Crypt password
	mailbox.Password, err = PasswdMD5Crypt(ctx, mailbox.Password)

	if err != nil {
		err = fmt.Errorf("Generate password md5-crypt failed: %w", err)
		return
	}

	mailbox.Username = strings.ToLower(mailbox.Username)
	mailbox.LocalPart = strings.ToLower(mailbox.LocalPart)
	mailbox.Domain = strings.ToLower(mailbox.Domain)

	now := time.Now().Unix()
	mailbox.CreateTime = now
	mailbox.UpdateTime = now
	mailbox.Active = 1
	mailbox.Maildir = fmt.Sprintf("%s@%s/", mailbox.LocalPart, mailbox.Domain)

	_, err = g.DB().Model("mailbox").Ctx(ctx).Insert(mailbox)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return fmt.Errorf("mailbox %s already exists", mailbox.Username)
		}
		return err
	}
	// maildirsize
	if e2 := ensureMaildirAndQuotaFile(ctx, mailbox); e2 != nil {
		g.Log().Warning(ctx, "ensureMaildirAndQuotaFile failed", e2)
	}
	return nil
}

func Update(ctx context.Context, mailbox *v1.Mailbox) (err error) {
	mailbox.UpdateTime = time.Now().Unix()
	if mailbox.Password != "" {
		mailbox.PasswordEncode = PasswdEncode(ctx, mailbox.Password)
		mailbox.Password, err = PasswdMD5Crypt(ctx, mailbox.Password)

		if err != nil {
			err = fmt.Errorf("Generate password md5-crypt failed: %w", err)
			return
		}
	}

	mailbox.Username = strings.ToLower(mailbox.Username)
	mailbox.LocalPart = strings.ToLower(mailbox.LocalPart)
	mailbox.Domain = strings.ToLower(mailbox.Domain)
	mailbox.Maildir = fmt.Sprintf("%s@%s/", mailbox.LocalPart, mailbox.Domain)

	m := gconv.Map(mailbox)
	delete(m, "create_time")
	delete(m, "used_quota")

	var mb v1.Mailbox
	err = g.DB().Model("mailbox").Where("username", mailbox.Username).Scan(&mb)

	_, err = g.DB().Model("mailbox").
		Ctx(ctx).
		Where("username", mailbox.Username).
		Update(m)
	if err != nil {

		return err
	}

	if mailbox.QuotaActive == 1 && mailbox.Quota != 0 {
		if mailbox.Quota != mb.Quota {
			if e2 := updateMaildirQuotaHeader(ctx, mailbox); e2 != nil {
				g.Log().Warning(ctx, "updateMaildirQuotaHeader failed", e2)
			}
		}
	}
	return nil
}

func Delete(ctx context.Context, email string) error {
	_, err := g.DB().Model("mailbox").
		Ctx(ctx).
		Where("username", email).
		Delete()
	return err
}

func DeleteBatch(ctx context.Context, emails []string) (int64, error) {
	if len(emails) == 0 {
		return 0, nil
	}

	result, err := g.DB().Model("mailbox").
		Ctx(ctx).
		WhereIn("username", emails).
		Delete()

	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func Get(ctx context.Context, domain, keyword string, page, pageSize int) ([]v1.Mailbox, int, error) {
	m := g.DB().Model("mailbox").Order("create_time", "desc")

	if domain != "" {
		m.Where("domain", domain)
	}

	if keyword != "" {
		m = m.WhereLike("username", fmt.Sprintf("%%%s%%", keyword))
	}

	count, err := m.Count()
	if err != nil {
		return nil, 0, err
	}

	var mailboxes []v1.Mailbox
	err = m.Page(page, pageSize).Scan(&mailboxes)

	return mailboxes, count, err
}

func All(ctx context.Context, domain string) ([]v1.Mailbox, error) {
	var mailboxes []v1.Mailbox
	query := g.DB().Model("mailbox")

	if domain != "" {
		query.Where("domain", domain)
	}

	err := query.Scan(&mailboxes)
	if err != nil {
		return nil, err
	}
	return mailboxes, nil
}

func AllEmail(ctx context.Context, domain string) ([]string, error) {
	var emails []string

	query := g.DB().Model("mailbox")

	if domain != "" {
		query.Where("domain", domain)
	}

	arr, err := query.Array("username")
	if err != nil {
		return nil, err
	}

	for _, v := range arr {
		emails = append(emails, v.String())
	}

	return emails, nil
}

func PasswordByEmail(ctx context.Context, email string) (pwd string, err error) {
	val, err := g.DB().Model("mailbox").Where("username", email).Value("password_encode")

	if err != nil {
		return
	}

	return PasswdDecode(ctx, val.String())
}

func PasswdEncode(ctx context.Context, password string) string {
	return hex.EncodeToString([]byte(base64.StdEncoding.EncodeToString([]byte(password))))
}

func PasswdDecode(ctx context.Context, password string) (string, error) {
	decoded, err := hex.DecodeString(password)
	if err != nil {
		return "", err
	}
	decodedStr, err := base64.StdEncoding.DecodeString(string(decoded))
	if err != nil {
		return "", err
	}
	return string(decodedStr), nil
}

func PasswdMD5Crypt(ctx context.Context, password string) (string, error) {
	// use md5_crypt package to generate MD5-CRYPT hash
	crypter := md5_crypt.New()

	// Generate generates a hash using the given password and salt.
	// The salt is optional and can be empty.
	// If the salt is empty, a random salt will be generated.
	result, err := crypter.Generate([]byte(password), []byte(""))
	if err != nil {
		return "", err
	}

	return result, nil
}

func generateRandomPassword(charset string, length int) string {
	if charset == "" {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}
func BatchAdd(ctx context.Context, domain string, quota int, count int, prefix string, quotaActive int) ([]string, error) {

	if prefix == "" {
		randomPre := make([]byte, 4)
		for j := 0; j < 4; j++ {
			randomPre[j] = byte(rand.Intn(26) + 97) // a-z的ASCII码
		}

		prefix = string(randomPre)
	}

	matched, err := regexp.MatchString(`^[\w-]+$`, prefix)
	if err != nil {
		return nil, fmt.Errorf("Prefix validation error: %w", err)
	}
	if !matched {
		return nil, fmt.Errorf("Prefixes can contain only letters, numbers, underscores, and hyphens")
	}

	rand.Seed(time.Now().UnixNano())

	createdEmails := make([]string, 0, count)

	timestamp := time.Now().Unix()

	//passwordEncoded := PasswdEncode(ctx, password)
	//passwordCrypted, err := PasswdMD5Crypt(ctx, password)
	//if err != nil {
	//	return nil, fmt.Errorf("Generate password md5-crypt failed: %w", err)
	//}

	mailboxes := make([]v1.Mailbox, 0, count)
	emailList := make([]string, 0, count)

	for i := 0; i < count; i++ {
		// 生成本地部分（用户名）: prefix + 自增数字  + 随机2位字母
		randomChars := make([]byte, 2)
		for j := 0; j < 2; j++ {
			randomChars[j] = byte(rand.Intn(26) + 97) // a-z的ASCII码
		}

		localPart := fmt.Sprintf("%s%d%s", prefix, i, string(randomChars))
		username := localPart + "@" + domain

		password := generateRandomPassword("", 8)

		passwordEncoded := PasswdEncode(ctx, password)
		passwordCrypted, _ := PasswdMD5Crypt(ctx, password)

		mailbox := v1.Mailbox{
			Username:       username,
			Password:       passwordCrypted,
			PasswordEncode: passwordEncoded,
			FullName:       localPart,
			IsAdmin:        0,
			Quota:          int64(quota),
			LocalPart:      localPart,
			Domain:         domain,
			CreateTime:     timestamp,
			UpdateTime:     timestamp,
			Active:         1,
			Maildir:        fmt.Sprintf("%s@%s/", localPart, domain),
			QuotaActive:    quotaActive,
		}

		mailboxes = append(mailboxes, mailbox)
		emailList = append(emailList, username)
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

		//domainExists, err := tx.Model("domain").Where("domain", domain).WhereNot("active", 0).One()
		//if err != nil {
		//	return fmt.Errorf("Check the domain for a failure: %w", err)
		//}
		//if domainExists.IsEmpty() {
		//	return fmt.Errorf(" %s Not present or activated", domain)
		//}

		const batchSize = 100
		for i := 0; i < len(mailboxes); i += batchSize {
			end := i + batchSize
			if end > len(mailboxes) {
				end = len(mailboxes)
			}

			batch := mailboxes[i:end]

			_, err := tx.Model("mailbox").Ctx(ctx).InsertIgnore(batch)
			if err != nil {
				return fmt.Errorf("Batch insert mailbox failed (batch %d-%d): %w", i, end-1, err)
			}

			for j := i; j < end; j++ {
				createdEmails = append(createdEmails, emailList[j])
			}

		}

		// If the limit is enabled, asynchronous will initialize the maildir and maildirsize for batch-created email accounts, and limit the concurrency
		if quotaActive != 1 || quota <= 0 {
			return nil
		}

		go func(mbs []v1.Mailbox) {
			sem := make(chan struct{}, 10)
			for i := range mbs {
				mb := mbs[i]
				sem <- struct{}{}
				go func(mb v1.Mailbox) {
					defer func() { <-sem }()
					_ = ensureMaildirAndQuotaFile(context.Background(), &mb)
				}(mb)
			}

			for i := 0; i < cap(sem); i++ {
				sem <- struct{}{}
			}
		}(mailboxes)

		return nil
	})

	if err != nil {

		return nil, fmt.Errorf("Failed to create email: %w", err)
	}

	if len(emailList) == 0 {
		return nil, fmt.Errorf("Failed to create any mailbox")
	}

	return emailList, nil
}

// AddImport
func AddImport(ctx context.Context, mailbox *v1.Mailbox) (err error) {

	if mailbox.PasswordEncode != "" {

		if mailbox.Password == "" {
			mailbox.Password, err = PasswdDecode(ctx, mailbox.PasswordEncode)
			if err != nil {
				err = fmt.Errorf("Decode password failed: %w", err)
				return
			}

			mailbox.Password, err = PasswdMD5Crypt(ctx, mailbox.Password)
			if err != nil {
				err = fmt.Errorf("Generate password md5-crypt failed: %w", err)
				return
			}
		}

	} else {

		mailbox.PasswordEncode = PasswdEncode(ctx, mailbox.Password)
		mailbox.Password, err = PasswdMD5Crypt(ctx, mailbox.Password)
		if err != nil {
			err = fmt.Errorf("Generate password md5-crypt failed: %w", err)
			return
		}
	}

	mailbox.Username = strings.ToLower(mailbox.Username)
	if mailbox.LocalPart == "" && mailbox.Username != "" {
		if parts := strings.Split(mailbox.Username, "@"); len(parts) == 2 {
			mailbox.LocalPart = parts[0]
		}
	}
	mailbox.LocalPart = strings.ToLower(mailbox.LocalPart)
	mailbox.Domain = strings.ToLower(mailbox.Domain)

	now := time.Now().Unix()
	mailbox.CreateTime = now
	mailbox.UpdateTime = now
	mailbox.Active = 1
	mailbox.QuotaActive = 1
	mailbox.Maildir = fmt.Sprintf("%s@%s/", mailbox.LocalPart, mailbox.Domain)

	_, err = g.DB().Model("mailbox").Ctx(ctx).InsertIgnore(mailbox)
	if err != nil {
		return err
	}

	if e2 := ensureMaildirAndQuotaFile(ctx, mailbox); e2 != nil {
		g.Log().Warning(ctx, "AddImport ensureMaildirAndQuotaFile failed", e2)
	}
	return nil
}

// NormalizeMailboxes normalizes mailbox usernames, local parts, domains, and maildirs to lowercase.
func NormalizeMailboxes() (err error) {
	// Attempt update mailboxes with uppercase letters in username
	_, err = g.DB().Model("mailbox").Where("username ~ '[A-Z]+'").Update(g.Map{
		"username":   gdb.Raw("LOWER(username)"),
		"local_part": gdb.Raw("LOWER(local_part)"),
		"domain":     gdb.Raw("LOWER(domain)"),
		"maildir":    gdb.Raw("LOWER(maildir)"),
	})

	return
}

func maildirRoot(m *v1.Mailbox) string {
	vmailRoot := public.AbsPath("../vmail-data")
	return filepath.Join(vmailRoot, m.Domain, m.LocalPart)
}

func atomicWriteMaildirsize(path string, content []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, 0644); err != nil {
		return err
	}
	_ = os.Chown(tmp, 150, 8)
	return os.Rename(tmp, path)
}

func ensureMaildirAndQuotaFile(ctx context.Context, m *v1.Mailbox) error {
	root := maildirRoot(m)
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}

	if err := public.ChownDovecot(root); err != nil {
		g.Log().Warning(ctx, "chown userDir failed", root, err)

	}

	fp := filepath.Join(root, "maildirsize")

	quotaVal := int64(0)
	if m.QuotaActive == 1 && m.Quota > 0 {
		quotaVal = m.Quota
	}
	content := fmt.Sprintf("%dS\n0 0\n", quotaVal)
	if _, err := os.Stat(fp); err == nil {

		return updateMaildirQuotaHeader(ctx, m)
	}
	if err := atomicWriteMaildirsize(fp, []byte(content)); err != nil {

		return err
	}
	recalcQuotaAsync(ctx, m.Username)
	return nil
}

func updateMaildirQuotaHeader(ctx context.Context, m *v1.Mailbox) error {
	root := maildirRoot(m)
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}
	fp := filepath.Join(root, "maildirsize")
	quotaVal := int64(0)
	if m.QuotaActive == 1 && m.Quota > 0 {
		quotaVal = m.Quota
	}
	header := fmt.Sprintf("%dS\n", quotaVal)

	data, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			if err2 := atomicWriteMaildirsize(fp, []byte(header+"0 0\n")); err2 != nil {
				return err2
			}
			recalcQuotaAsync(ctx, m.Username)
			return nil
		}
		return err
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		if err2 := atomicWriteMaildirsize(fp, []byte(header+"0 0\n")); err2 != nil {
			return err2
		}
		recalcQuotaAsync(ctx, m.Username)
		return nil
	}
	usageLines := make([]string, 0, len(lines))
	for i := 1; i < len(lines); i++ {
		l := strings.TrimSpace(lines[i])
		if l != "" {
			//   "<bytes> <count>"
			parts := strings.Fields(l)
			if len(parts) >= 1 {
				usageLines = append(usageLines, l)
			}
		}
	}
	if len(usageLines) == 0 {
		usageLines = append(usageLines, "0 0")
	}
	newContent := header + strings.Join(usageLines, "\n") + "\n"
	if err := atomicWriteMaildirsize(fp, []byte(newContent)); err != nil {
		return err
	}
	recalcQuotaAsync(ctx, m.Username)
	return nil
}

// doveadm quota recalc -u <user>
func recalcQuotaAsync(ctx context.Context, username string) {
	if username == "" {
		return
	}
	go func(u string) {
		dk, err := docker.NewDockerAPI()
		if err != nil {
			return
		}
		defer dk.Close()

		_, _ = dk.ExecCommandByName(ctx, consts.SERVICES.Dovecot, []string{"doveadm", "quota", "recalc", "-u", u}, "root")
	}(username)
}
