package db

import (
	"database/sql"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

func EnsureAutoIncrement(db *gorm.DB, table, column string, start int) error {
	driver := db.Dialector.Name()

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	switch driver {
	case "postgres":
		var seqName sql.NullString
		if err := sqlDB.QueryRow(`SELECT pg_get_serial_sequence($1, $2)`, table, column).Scan(&seqName); err != nil {
			return fmt.Errorf("get sequence name: %w", err)
		}
		if !seqName.Valid {
			return fmt.Errorf("no sequence found for %s.%s", table, column)
		}

		var lastVal int64
		if err := sqlDB.QueryRow(`SELECT last_value FROM ` + seqName.String).Scan(&lastVal); err != nil {
			return fmt.Errorf("get last_value: %w", err)
		}

		if lastVal < int64(start) {
			if err := db.Exec(`ALTER SEQUENCE ` + seqName.String + ` RESTART WITH ` + strconv.Itoa(start)).Error; err != nil {
				return fmt.Errorf("restart sequence: %w", err)
			}
		}

	case "mysql":
		var autoInc sql.NullInt64
		q := `SELECT AUTO_INCREMENT
		      FROM information_schema.tables
		      WHERE table_name = ? AND table_schema = DATABASE()`
		if err := sqlDB.QueryRow(q, table).Scan(&autoInc); err != nil {
			return fmt.Errorf("get AUTO_INCREMENT: %w", err)
		}

		if autoInc.Valid && autoInc.Int64 < int64(start) {
			if err := db.Exec(`ALTER TABLE ` + table + ` AUTO_INCREMENT = ` + strconv.Itoa(start)).Error; err != nil {
				return fmt.Errorf("reset AUTO_INCREMENT: %w", err)
			}
		}

	default:
		return fmt.Errorf("unsupported driver: %s", driver)
	}
	return nil
}
