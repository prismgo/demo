package schemademo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/prismgo/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// withIsolatedSQLite runs fn against a throwaway SQLite database.
//
// The database lives in a temporary directory and is removed afterwards, so
// destructive scenarios never touch the application database. SQLite schema
// changes have to run serially on a pinned connection, so the pool is capped
// at one connection for the lifetime of the scenario.
func withIsolatedSQLite(fn func(*gorm.DB) (string, error)) (string, error) {
	dir, err := os.MkdirTemp("", "prismgo-schema-demo-")
	if err != nil {
		return "", fmt.Errorf("create isolated sqlite directory: %w", err)
	}
	db, err := gorm.Open(
		sqlite.Open(filepath.Join(dir, "schema.sqlite")),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("open isolated sqlite database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("resolve isolated sqlite connection: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	value, runErr := fn(db)
	closeErr := sqlDB.Close()
	removeErr := os.RemoveAll(dir)
	if runErr != nil {
		return "", runErr
	}
	if closeErr != nil {
		return "", fmt.Errorf("close isolated sqlite connection: %w", closeErr)
	}
	if removeErr != nil {
		return "", fmt.Errorf("remove isolated sqlite directory: %w", removeErr)
	}
	return value, nil
}
