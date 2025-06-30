package database_initialization

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
)

// Automatically adding table fields
func AddColumnIfNotExists(table, column, columnType, defaultValue string, notNull bool) error {
	notNullStr := ""
	if notNull {
		notNullStr = "NOT NULL"
	}
	defaultStr := ""
	if defaultValue != "" {
		defaultStr = "DEFAULT " + defaultValue
	}
	sql := fmt.Sprintf(`
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = '%s' AND column_name = '%s'
	) THEN
		ALTER TABLE %s ADD COLUMN %s %s %s %s;
	END IF;
END $$;`, table, column, table, column, columnType, notNullStr, defaultStr)
	_, err := g.DB().Exec(context.Background(), sql)
	if err != nil {
		g.Log().Error(context.Background(), "Failed to add column:", err, sql)
	}
	return err
}
