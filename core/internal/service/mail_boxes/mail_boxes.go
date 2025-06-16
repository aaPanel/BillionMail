package mail_boxes

import (
	v1 "billionmail-core/api/mail_boxes/v1"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/GehirnInc/crypt/md5_crypt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"math/rand"
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
	return err
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

	_, err = g.DB().Model("mailbox").
		Ctx(ctx).
		Where("username", mailbox.Username).
		Update(m)
	return err
}

func Delete(ctx context.Context, email string) error {
	_, err := g.DB().Model("mailbox").
		Ctx(ctx).
		Where("username", email).
		Delete()
	return err
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

func BatchAdd(ctx context.Context, domain, password string, quota int, count int, prefix string) ([]string, error) {

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

	passwordEncoded := PasswdEncode(ctx, password)
	passwordCrypted, err := PasswdMD5Crypt(ctx, password)
	if err != nil {
		return nil, fmt.Errorf("Generate password md5-crypt failed: %w", err)
	}

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

		return nil
	})

	if err != nil {

		return nil, fmt.Errorf("Failed to create email: %w", err)
	}

	if len(createdEmails) == 0 {
		return nil, fmt.Errorf("Failed to create any mailbox")
	}

	return createdEmails, nil
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
