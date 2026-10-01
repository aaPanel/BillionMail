package database_initialization

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
	"time"
)

func init() {
	registerHandler(func() {

		ctx := context.Background()

		sqlList := []string{

			`-- Relay configuration 
			CREATE TABLE IF NOT EXISTS bm_relay_config (
			id BIGSERIAL PRIMARY KEY,
			remark varchar(255),  
			rtype varchar(30),  -- Relay type: gmail, sendgrid, custom, aws, mailgun, local
			relay_host varchar(255) NOT NULL, -- Relay server address 
			relay_port varchar(10) NOT NULL, -- Relay server port 
			auth_user varchar(255) NOT NULL, -- SMTP auth username
			auth_password varchar(255) NOT NULL, -- SMTP auth password (encrypted)
			ip varchar(255), -- IP for reminding user to update SPF record (optional) 
			host varchar(255), --  Hostname for reminding user to update SPF record (optional)
			helo_name varchar(255), -- HELO hostname 
			skip_tls_verify smallint DEFAULT 0, -- kip TLS verification: 1-skip, 0-do not skip 
			smtp_name varchar(50), --  SMTP server name
			active smallint NOT NULL DEFAULT 1, 
			auth_method varchar(20), -- Authentication method: LOGIN, PLAIN, CRAM-MD5, NONE
			create_time int NOT NULL default 0,
			update_time int NOT NULL default 0
			) `,

			`-- Relay to domain mapping
			CREATE TABLE IF NOT EXISTS bm_relay_domain_mapping (
			id BIGSERIAL PRIMARY KEY,
			relay_id BIGINT NOT NULL, --  Reference to id in bm_relay_config  
			sender_domain varchar(255) NOT NULL, -- Sender domain, e.g. "example.com"
			create_time int NOT NULL default 0
			) `,
			// 域名唯一约束
			`-- Domain to SMTP service name mapping table
			CREATE TABLE IF NOT EXISTS  bm_domain_smtp_transport(
			id BIGSERIAL PRIMARY KEY,
			atype varchar(50) NOT NULL DEFAULT 'relay', -- Transport type,  "relay"   "dedicated_ip"
			domain varchar(255) NOT NULL,  -- Sender domain, e.g. "@example.com"
			smtp_name varchar(255) NOT NULL, -- SMTP service name, e.g. "smtp_Relay_mail_xxxx_cn_gwt3_6_pvz_6"
			unique(domain)
			) `,

			`CREATE INDEX IF NOT EXISTS idx_domain_smtp_transport_domain ON bm_domain_smtp_transport(domain);`,
			`CREATE INDEX IF NOT EXISTS idx_domain_smtp_transport_atype ON bm_domain_smtp_transport(atype);`,
			`CREATE INDEX IF NOT EXISTS idx_domain_smtp_transport_domain_smtpName ON bm_domain_smtp_transport(domain,smtp_name);`,
			`CREATE INDEX IF NOT EXISTS idx_relay_domain_mapping_domain ON bm_relay_domain_mapping(sender_domain);`,
			`CREATE INDEX IF NOT EXISTS idx_relay_domain_mapping_relay_id ON bm_relay_domain_mapping(relay_id);`,
		}

		for _, sql := range sqlList {
			_, err := g.DB().Exec(ctx, sql)
			if err != nil {
				g.Log().Error(ctx, "Failed to create mail server tables:", err)
				return
			}
		}

		if err := cleanupAndIndexRelayDomainMapping(ctx); err != nil {
			g.Log().Error(ctx, "Failed to cleanup duplicates or create unique index:", err)
			return
		}

		// There is a rename table that has been synchronized. Skip.
		existsOld := checkTableExists(ctx, "bm_relay_old")
		if existsOld {
			return
		}

		//  Migrate the old data to the new structure
		if err := migrateRelayData(ctx); err != nil {
			g.Log().Error(ctx, "Failed to migrate relay data:", err)
			return
		}

		// Rename the old table (only after the migration is successful)
		if err := RenameRelayTable(ctx); err != nil {
			g.Log().Error(ctx, "Failed to rename old table bm_relay:", err)
			return
		}
		g.Log().Info(ctx, "Relay configuration migration completed successfully.")

	})
}

// cleanupAndIndexRelayDomainMapping
func cleanupAndIndexRelayDomainMapping(ctx context.Context) error {
	if !checkTableExists(ctx, "bm_relay_domain_mapping") {
		return gerror.New("Table bm_relay_domain_mapping does not exist")
	}

	exists, err := g.DB().Model("pg_indexes").Fields("1").
		Where("indexname", "uk_relay_domain").
		Value()
	if err != nil {
		return err
	}
	if exists != nil && exists.Int() == 1 {
		g.Log().Info(ctx, "Unique index uk_relay_domain already exists, skip cleanup.")
		return nil
	}

	_, err = g.DB().Exec(ctx, `
		DELETE FROM bm_relay_domain_mapping 
		WHERE id NOT IN (
			SELECT MIN(id)
			FROM bm_relay_domain_mapping
			GROUP BY relay_id, sender_domain
		)
	`)
	if err != nil {
		return gerror.Wrap(err, "Failed to remove duplicates from bm_relay_domain_mapping")
	}

	_, err = g.DB().Exec(ctx, `
		CREATE UNIQUE INDEX uk_relay_domain 
		ON bm_relay_domain_mapping(relay_id, sender_domain)
	`)
	if err != nil {
		return gerror.Wrap(err, "Failed to create unique index uk_relay_domain")
	}

	g.Log().Info(ctx, "Cleanup duplicates and created unique index on bm_relay_domain_mapping")
	return nil
}

func migrateRelayData(ctx context.Context) error {

	existsOld := checkTableExists(ctx, "bm_relay")
	if !existsOld {
		g.Log().Info(ctx, "Old table bm_relay does not exist, skip migration")
		return nil
	}

	existsNew := checkTableExists(ctx, "bm_relay_config")
	if !existsNew {
		return gerror.New("New table bm_relay_config does not exist, please create tables first")
	}

	newCount, err := g.DB().Model("bm_relay_config").Count()
	if err != nil {
		return gerror.Wrap(err, "Failed to query record count of new table")
	}
	if newCount > 0 {
		g.Log().Debugf(ctx, "bm_relay_config already has data (%d records), skip migration to prevent duplication", newCount)
		return nil
	}

	// 查询旧表数据
	var oldRelays []struct {
		Id            int64   `json:"id"`
		Remark        *string `json:"remark"`
		Rtype         *string `json:"rtype"`
		SenderDomain  *string `json:"sender_domain"`
		RelayHost     *string `json:"relay_host"`
		RelayPort     *string `json:"relay_port"`
		AuthUser      *string `json:"auth_user"`
		AuthPassword  *string `json:"auth_password"`
		Ip            *string `json:"ip"`
		Host          *string `json:"host"`
		HeloName      *string `json:"helo_name"`
		SkipTlsVerify *int    `json:"skip_tls_verify"`
		SmtpName      *string `json:"smtp_name"`
		Active        *int    `json:"active"`
		AuthMethod    *string `json:"auth_method"`
		CreateTime    int     `json:"create_time"`
		UpdateTime    int     `json:"update_time"`
	}

	err = g.DB().Model("bm_relay").
		Fields("id, remark, rtype, sender_domain, relay_host, relay_port, auth_user, auth_password, ip, host, helo_name, skip_tls_verify, smtp_name, active, auth_method, create_time, update_time").
		Scan(&oldRelays)
	if err != nil {
		return gerror.Wrap(err, "Failed to query old table bm_relay")
	}

	if len(oldRelays) == 0 {
		g.Log().Info(ctx, "No data in old table bm_relay, no migration needed")
		return nil
	}

	g.Log().Infof(ctx, "Start migrating %d relay configurations", len(oldRelays))

	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return gerror.Wrap(err, "Failed to begin transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			g.Log().Error(ctx, "Transaction rolled back during data migration")
		}
	}()

	for _, r := range oldRelays {

		configData := make(gdb.Map)

		if r.Remark != nil {
			configData["remark"] = *r.Remark
		}
		if r.Rtype != nil {
			configData["rtype"] = *r.Rtype
		}

		if r.RelayHost == nil || *r.RelayHost == "" {
			g.Log().Errorf(ctx, "Skip record ID=%d: relay_host is empty", r.Id)
			continue
		}
		configData["relay_host"] = *r.RelayHost

		if r.RelayPort == nil || *r.RelayPort == "" {
			configData["relay_port"] = "25"
		} else {
			configData["relay_port"] = *r.RelayPort
		}

		if r.AuthUser != nil {
			configData["auth_user"] = *r.AuthUser
		}
		if r.AuthPassword != nil {
			configData["auth_password"] = *r.AuthPassword
		}
		if r.Ip != nil {
			configData["ip"] = *r.Ip
		}
		if r.Host != nil {
			configData["host"] = *r.Host
		}
		if r.HeloName != nil {
			configData["helo_name"] = *r.HeloName
		}
		if r.SkipTlsVerify != nil {
			configData["skip_tls_verify"] = *r.SkipTlsVerify
		} else {
			configData["skip_tls_verify"] = 0
		}

		if r.SmtpName != nil && *r.SmtpName != "" {
			configData["smtp_name"] = *r.SmtpName
		} else {
			// Generate unique name
			configData["smtp_name"] = "MigratedRelay_" + time.Now().Format("20060102_150405") + "_" + strings.ReplaceAll(*r.RelayHost, ".", "_")
		}

		if r.Active != nil {
			configData["active"] = *r.Active
		} else {
			configData["active"] = 1
		}

		if r.AuthMethod != nil {
			configData["auth_method"] = *r.AuthMethod
		} else {
			configData["auth_method"] = "LOGIN"
		}

		configData["create_time"] = r.CreateTime
		configData["update_time"] = r.UpdateTime

		// Insert into bm_relay_config
		result, err := tx.Model("bm_relay_config").Data(configData).Insert()
		if err != nil {
			return gerror.Wrapf(err, "Failed to insert bm_relay_config (ID=%d)", r.Id)
		}

		relayConfigId, err := result.LastInsertId()
		if err != nil {
			return gerror.Wrapf(err, "Failed to get last insert ID (ID=%d)", r.Id)
		}

		// Insert domain mapping
		if r.SenderDomain != nil && *r.SenderDomain != "" {
			mappingData := gdb.Map{
				"relay_id":      relayConfigId,
				"sender_domain": *r.SenderDomain,
				"create_time":   r.CreateTime,
			}
			_, err = tx.Model("bm_relay_domain_mapping").Data(mappingData).Insert()
			if err != nil {
				return gerror.Wrapf(err, "Failed to insert bm_relay_domain_mapping (ID=%d)", r.Id)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return gerror.Wrap(err, "Failed to commit transaction after data migration")
	}

	g.Log().Infof(ctx, "Data migration succeeded! Migrated %d relay configurations", len(oldRelays))
	return nil
}

func checkTableExists(ctx context.Context, tableName string) bool {
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = ?
	)`

	result, err := g.DB().GetOne(ctx, query, tableName)
	if err != nil {
		g.Log().Errorf(ctx, "Error checking if table %s exists: %v", tableName, err)
		return false
	}

	return result["exists"].Bool()
}

// RenameRelayTable Rename the old relay table
func RenameRelayTable(ctx context.Context) error {

	existsNew := checkTableExists(ctx, "bm_relay")
	if !existsNew {
		return gerror.New("Table bm_relay does not exist, no need to rename")
	}

	existsOld := checkTableExists(ctx, "bm_relay_old")
	if !existsOld {
		return gerror.New("Table bm_relay_old already exists, skipping rename operation")
	}

	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return gerror.Wrap(err, "Failed to begin transaction for renaming table")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			g.Log().Error(ctx, "Transaction rolled back during table rename")
		}
	}()

	_, err = tx.Exec("ALTER TABLE bm_relay RENAME TO bm_relay_old")

	if err != nil {
		return gerror.Wrap(err, "Failed to rename bm_relay to bm_relay_old")
	}

	if err = tx.Commit(); err != nil {
		return gerror.Wrap(err, "Failed to commit transaction after renaming table")
	}

	g.Log().Info(ctx, "Successfully renamed table bm_relay to bm_relay_old")
	return nil
}
