package commands

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func runPostgresMigration(ctx context.Context, executable, root, name string) (result Result, err error) {
	dsn := strings.TrimSpace(os.Getenv("PRISMGO_POSTGRES_TEST_DSN"))
	if dsn == "" {
		return Result{}, fmt.Errorf("%s requires PRISMGO_POSTGRES_TEST_DSN for PostgreSQL integration", name)
	}
	address, err := url.Parse(dsn)
	if err != nil {
		return Result{}, fmt.Errorf("parse %s PostgreSQL URL: %w", name, err)
	}
	if address.Scheme != "postgres" && address.Scheme != "postgresql" {
		return Result{}, fmt.Errorf("%s requires a PostgreSQL URL, got scheme %q", name, address.Scheme)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return Result{}, fmt.Errorf("open commands demo PostgreSQL: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return Result{}, fmt.Errorf("resolve commands demo PostgreSQL connection: %w", err)
	}
	defer sqlDB.Close()
	if err := sqlDB.PingContext(ctx); err != nil {
		return Result{}, fmt.Errorf("connect commands demo PostgreSQL: %w", err)
	}
	schema := fmt.Sprintf("prismgo_commands_%d", time.Now().UnixNano())
	if err := db.WithContext(ctx).Exec("CREATE SCHEMA " + schema).Error; err != nil {
		return Result{}, fmt.Errorf("create commands demo PostgreSQL schema: %w", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if cleanupErr := db.WithContext(cleanupCtx).Exec("DROP SCHEMA " + schema + " CASCADE").Error; cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("clean commands demo PostgreSQL schema: %w", cleanupErr))
		}
	}()
	if err := db.WithContext(ctx).Exec("CREATE TYPE " + schema + ".commands_demo_status AS ENUM ('pending', 'done')").Error; err != nil {
		return Result{}, fmt.Errorf("create commands demo PostgreSQL enum: %w", err)
	}
	if err := db.WithContext(ctx).Exec("CREATE TABLE " + schema + ".commands_demo_extra (id BIGINT PRIMARY KEY)").Error; err != nil {
		return Result{}, fmt.Errorf("create commands demo PostgreSQL table: %w", err)
	}
	path := filepath.Join(root, "database", "empty")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return Result{}, fmt.Errorf("prepare empty PostgreSQL migration directory: %w", err)
	}
	query := address.Query()
	query.Set("search_path", schema)
	address.RawQuery = query.Encode()
	args := []string{"migrate:fresh", "--drop-types", "--path=database/empty"}
	output, err := invoke(ctx, executable, root, []string{"DB_CONNECTION=postgres", "POSTGRES_DSN=" + address.String()}, args...)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", name, err)
	}
	var enumCount, tableCount int64
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE n.nspname = ? AND t.typname = 'commands_demo_status'", schema).Scan(&enumCount).Error; err != nil {
		return Result{}, fmt.Errorf("check commands demo PostgreSQL enum: %w", err)
	}
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'commands_demo_extra'", schema).Scan(&tableCount).Error; err != nil {
		return Result{}, fmt.Errorf("check commands demo PostgreSQL table: %w", err)
	}
	if enumCount != 0 || tableCount != 0 {
		return Result{}, fmt.Errorf("%s remaining enum/table = %d/%d, want 0/0; output = %q", name, enumCount, tableCount, output)
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: strings.TrimSpace(output) + "\nverified PostgreSQL enum and table removed"}, nil
}
