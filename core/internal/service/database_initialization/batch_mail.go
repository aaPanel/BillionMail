package database_initialization

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	registerHandler(func() {
		batchMailSQLList := []string{

			`CREATE TABLE IF NOT EXISTS bm_contact_groups (
                id SERIAL PRIMARY KEY,
                name VARCHAR(255) NOT NULL,
                description TEXT,
                create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                update_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                UNIQUE(name)
            )`,

			`CREATE TABLE IF NOT EXISTS bm_contacts (
                id SERIAL PRIMARY KEY,
                email VARCHAR(320) NOT NULL,
                group_id INTEGER,
                active INTEGER DEFAULT 1,
                task_id INTEGER DEFAULT 0,
                create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                status INTEGER DEFAULT 0, -- 0: Unconfirmed, 1: Confirmed
                attribs   JSONB DEFAULT '{}'::jsonb,
                FOREIGN KEY (group_id) REFERENCES bm_contact_groups(id) ON DELETE SET NULL,
                UNIQUE(group_id, email)
            )`,

			`CREATE TABLE IF NOT EXISTS email_templates (
                id SERIAL PRIMARY KEY,
                temp_name VARCHAR(255),
                add_type SMALLINT NOT NULL DEFAULT 0,
                content TEXT NOT NULL DEFAULT '',
                render TEXT NOT NULL DEFAULT '',
                create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                update_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
    			UNIQUE(temp_name)
            )`,

			`CREATE TABLE IF NOT EXISTS email_tasks (
                id SERIAL PRIMARY KEY,
                task_name VARCHAR(255) NOT NULL,
                addresser VARCHAR(320) NOT NULL,
                subject TEXT,
                full_name VARCHAR(255),
                remark TEXT,
                recipient_count INTEGER NOT NULL,
                task_process SMALLINT NOT NULL DEFAULT 0,
                pause SMALLINT NOT NULL DEFAULT 0,
                template_id INTEGER NOT NULL,
                is_record SMALLINT NOT NULL DEFAULT 0,
                unsubscribe SMALLINT NOT NULL DEFAULT 0,
                threads INTEGER NOT NULL DEFAULT 0,
                etypes VARCHAR(320) NOT NULL DEFAULT '1',
                track_open SMALLINT NOT NULL DEFAULT 1,
                track_click SMALLINT NOT NULL DEFAULT 1,
                start_time INTEGER NOT NULL DEFAULT 0,
                create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                update_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                active SMALLINT NOT NULL DEFAULT 0,
    			add_type SMALLINT NOT NULL DEFAULT 0,	
                FOREIGN KEY (template_id) REFERENCES email_templates(id)
            )`,

			`CREATE TABLE IF NOT EXISTS recipient_info (
                id SERIAL PRIMARY KEY,
                task_id INTEGER NOT NULL,
                recipient VARCHAR(320) NOT NULL,
                is_sent SMALLINT DEFAULT 0,
                sent_time INTEGER NOT NULL DEFAULT 0,
                message_id TEXT NOT NULL,
                create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                FOREIGN KEY (task_id) REFERENCES email_tasks(id) ON DELETE CASCADE,
                UNIQUE(task_id, recipient)
            )`,

			`CREATE TABLE IF NOT EXISTS unsubscribe_records (
                id SERIAL PRIMARY KEY,
                email VARCHAR(320) NOT NULL,
                group_id INTEGER,
                template_id INTEGER,
                task_id INTEGER,
                unsubscribe_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                FOREIGN KEY (group_id) REFERENCES bm_contact_groups(id) ON DELETE SET NULL,
                FOREIGN KEY (template_id) REFERENCES email_templates(id) ON DELETE SET NULL,
                FOREIGN KEY (task_id) REFERENCES email_tasks(id) ON DELETE SET NULL
            )`,

			`CREATE TABLE IF NOT EXISTS abnormal_recipient (
                id SERIAL PRIMARY KEY,
                create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                recipient VARCHAR(320) NOT NULL,
    			count INTEGER NOT NULL,
    			description VARCHAR(255) NOT NULL,   
    			add_type SMALLINT NOT NULL DEFAULT 0, 
    			UNIQUE(recipient)
            )`,

			`CREATE TABLE IF NOT EXISTS api_templates (
                id SERIAL PRIMARY KEY,
                api_key VARCHAR(64) NOT NULL,
                api_name VARCHAR(255) NOT NULL,
                template_id INTEGER NOT NULL,
                subject TEXT NOT NULL,
                addresser VARCHAR(320) NOT NULL,
                full_name VARCHAR(255),
                unsubscribe SMALLINT NOT NULL DEFAULT 0,
                track_open SMALLINT NOT NULL DEFAULT 1,
                track_click SMALLINT NOT NULL DEFAULT 1,
                active SMALLINT NOT NULL DEFAULT 1,
                create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                update_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                expire_time INTEGER DEFAULT 0,
                last_key_update_time INTEGER DEFAULT 0,
                UNIQUE(api_key)
            )`,

			`CREATE TABLE IF NOT EXISTS api_mail_logs (
                id SERIAL PRIMARY KEY,
                api_id INTEGER NOT NULL,
                recipient VARCHAR(320) NOT NULL,
                message_id TEXT NOT NULL,
                addresser VARCHAR(320) NOT NULL,
                status SMALLINT NOT NULL DEFAULT 0, -- 0:to send, 2:send, 3:send failed
                error_message TEXT, 
                send_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())

            )`,

			`CREATE TABLE IF NOT EXISTS api_ip_whitelist (
    				id SERIAL PRIMARY KEY,
    				api_id INTEGER NOT NULL,
    				ip VARCHAR(45) NOT NULL,
    				create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
    				FOREIGN KEY (api_id) REFERENCES api_templates(id) ON DELETE CASCADE
			)`,

			`CREATE INDEX IF NOT EXISTS idx_unsubscribe_email ON unsubscribe_records (email)`,
			`CREATE INDEX IF NOT EXISTS idx_unsubscribe_time ON unsubscribe_records (unsubscribe_time)`,
			`CREATE INDEX IF NOT EXISTS idx_bm_contacts_email ON bm_contacts (email)`,
			`CREATE INDEX IF NOT EXISTS idx_bm_contacts_active ON bm_contacts (active)`,
			`CREATE INDEX IF NOT EXISTS idx_abnormal_recipient_count ON abnormal_recipient (recipient,count)`,
			`CREATE INDEX IF NOT EXISTS idx_group_email ON bm_contacts(group_id, email)`,
			`CREATE INDEX IF NOT EXISTS idx_api_mail_logs_api_id ON api_mail_logs (api_id)`,
			`CREATE INDEX IF NOT EXISTS idx_api_mail_logs_recipient ON api_mail_logs (recipient)`,
			`CREATE INDEX IF NOT EXISTS idx_api_mail_logs_message_id ON api_mail_logs (message_id)`,
		}

		for _, sql := range batchMailSQLList {
			_, err := g.DB().Exec(context.Background(), sql)
			if err != nil {
				g.Log().Error(context.Background(), "Failed to execute batch mail SQL:", err, sql)
				return
			}
		}
		// bm_contacts  attribs
		_ = AddColumnIfNotExists("bm_contacts", "attribs", "JSONB", "'{}'::jsonb", false)
		//  bm_contacts status
		_ = AddColumnIfNotExists("bm_contacts", "status", "INTEGER", "0", false)
		//  api_mail_logs status
		_ = AddColumnIfNotExists("api_mail_logs", "status", "SMALLINT", "0", true)
		//  api_mail_logs   error_message
		_ = AddColumnIfNotExists("api_mail_logs", "error_message", "TEXT", "''", false)
		//  api_mail_logs  create_time
		_ = AddColumnIfNotExists("api_mail_logs", "create_time", "INTEGER", "EXTRACT(EPOCH FROM NOW())", true)

		g.Log().Info(context.Background(), "Batch mail tables initialized successfully")
	})
}
