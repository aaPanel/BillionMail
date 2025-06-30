package database_initialization

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	registerHandler(func() {
		row_sql_list := []string{
			`CREATE TABLE IF NOT EXISTS bm_sender_ip_warmup (
				id serial primary key,
				sender_ip text not null default '',
				period integer not null default 45, -- warmup period in days
				score integer not null default 40, -- 0-100
				last_evaluate_time integer not null default 0, -- last evaluate time
				progress integer not null default 0, -- progress of warmup 0-100
				re_warmups integer not null default 0, -- re-warmup times
				begin_time integer not null DEFAULT EXTRACT(EPOCH FROM NOW()), -- begin of warmup
				end_time integer not null default 0, -- end of warmup
				create_time integer DEFAULT EXTRACT(EPOCH FROM NOW()),
				update_time integer DEFAULT EXTRACT(EPOCH FROM NOW())
			)`,

			`CREATE INDEX IF NOT EXISTS idx_bm_sender_ip_warmup_sender_ip ON bm_sender_ip_warmup (sender_ip)`,

			`CREATE TABLE IF NOT EXISTS bm_sender_ip_mail_provider (
				id serial primary key,
				sender_ip text not null default '',
				mail_provider text not null default 'other', -- mail provider name
				score integer not null default 40, -- 0-100
				last_evaluate_time integer not null default 0, -- last evaluate time
				create_time integer not null default EXTRACT(EPOCH FROM NOW()),
				update_time integer not null default EXTRACT(EPOCH FROM NOW())
			)`,

			`CREATE INDEX IF NOT EXISTS idx_bm_sender_ip_mail_provider_sender_ip ON bm_sender_ip_mail_provider (sender_ip, mail_provider)`,

			`CREATE TABLE IF NOT EXISTS bm_campaign_warmup (
				id serial primary key,
				task_id bigint not null default 0, -- campaign task ID
				warmup_id bigint not null default 0 -- sender IP warmup ID
			)`,

			`CREATE UNIQUE INDEX IF NOT EXISTS idx_bm_campaign_warmup_task_id_warmup_id ON bm_campaign_warmup (task_id, warmup_id)`,
		}

		for _, sql := range row_sql_list {
			_, err := g.DB().Exec(context.Background(), sql)
			if err != nil {
				g.Log().Error(context.Background(), "Failed to create table:", err)
				return
			}
		}
	})
}
