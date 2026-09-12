package commands

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func runLocalMigration(ctx context.Context, executable, root, name string) (Result, error) {
	if err := prepareMigrationSources(root); err != nil {
		return Result{}, err
	}
	databasePath := filepath.Join(root, "commands.sqlite")
	environment := []string{"DB_CONNECTION=sqlite", "DB_DATABASE=" + databasePath}
	run := func(extraEnv []string, args ...string) (string, error) {
		output, err := invoke(ctx, executable, root, append(environment, extraEnv...), args...)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(output), nil
	}
	args := []string{"migrate", "--database=sqlite"}
	if name == "production-guard" {
		_, err := run([]string{"APP_ENV=production"}, args...)
		if err == nil || !strings.Contains(err.Error(), "rerun with --force") {
			return Result{}, fmt.Errorf("production migrate error = %v, want --force guard", err)
		}
		args = append(args, "--force")
	}
	if name == "sqlite-fresh" {
		if _, err := run(nil, "migrate", "--database=sqlite"); err != nil {
			return Result{}, fmt.Errorf("prepare SQLite migrations: %w", err)
		}
		if err := withSQLiteDemoDB(databasePath, func(db *gorm.DB) error {
			if err := db.Exec("CREATE TABLE commands_demo_extra (id INTEGER PRIMARY KEY)").Error; err != nil {
				return fmt.Errorf("prepare SQLite extra table: %w", err)
			}
			if err := db.Exec("CREATE VIEW commands_demo_view AS SELECT id FROM users").Error; err != nil {
				return fmt.Errorf("prepare SQLite view: %w", err)
			}
			return nil
		}); err != nil {
			return Result{}, err
		}
		args = []string{"migrate:fresh", "--database=sqlite", "--drop-views"}
	}
	extraEnv := []string(nil)
	if name == "production-guard" {
		extraEnv = []string{"APP_ENV=production"}
	}
	output, err := run(extraEnv, args...)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", name, err)
	}
	var count int64
	if err := withSQLiteDemoDB(databasePath, func(db *gorm.DB) error {
		if err := db.Raw("SELECT COUNT(*) FROM migrations").Scan(&count).Error; err != nil {
			return fmt.Errorf("%s count SQLite migrations: %w", name, err)
		}
		if count != int64(len(migrationSourceNames())) {
			return fmt.Errorf("%s SQLite migration rows = %d, want %d", name, count, len(migrationSourceNames()))
		}
		if name == "sqlite-fresh" {
			var extras int64
			if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE name IN ('commands_demo_extra', 'commands_demo_view')").Scan(&extras).Error; err != nil {
				return fmt.Errorf("inspect SQLite extra objects: %w", err)
			}
			if extras != 0 {
				return fmt.Errorf("sqlite-fresh extra objects = %d, want 0", extras)
			}
		}
		return nil
	}); err != nil {
		return Result{}, err
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: output}, nil
}

func withSQLiteDemoDB(path string, use func(*gorm.DB) error) (err error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open commands demo SQLite database: %w", err)
	}
	connection, err := db.DB()
	if err != nil {
		return fmt.Errorf("access commands demo SQLite connection: %w", err)
	}
	defer func() {
		if closeErr := connection.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close commands demo SQLite connection: %w", closeErr))
		}
	}()
	return use(db)
}
