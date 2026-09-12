package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

func migrationSourceNames() []string {
	return []string{
		"202606050000_create_users_table.go",
		"202609080000_create_demo_orders_table.go",
		"202609080010_create_demo_receipts_table.go",
	}
}

func runMySQLMigration(ctx context.Context, executable, root, name string) (result Result, err error) {
	dsn := strings.TrimSpace(os.Getenv("PRISMGO_COMMANDS_MYSQL_TEST_DSN"))
	if dsn == "" {
		return Result{}, fmt.Errorf("%s requires PRISMGO_COMMANDS_MYSQL_TEST_DSN for a dedicated MySQL test database", name)
	}
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		return Result{}, fmt.Errorf("parse commands demo MySQL DSN: %w", err)
	}
	if !strings.HasPrefix(config.DBName, "prismgo_commands_") {
		return Result{}, fmt.Errorf("commands demo MySQL database %q must use the prismgo_commands_ prefix", config.DBName)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return Result{}, fmt.Errorf("open commands demo MySQL database: %w", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return Result{}, fmt.Errorf("connect commands demo MySQL database: %w", err)
	}
	if err := clearDemoTables(ctx, db); err != nil {
		return Result{}, fmt.Errorf("prepare commands demo MySQL database: %w", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if cleanupErr := clearDemoTables(cleanupCtx, db); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("clean commands demo MySQL database: %w", cleanupErr))
		}
	}()
	if err := prepareMigrationSources(root); err != nil {
		return Result{}, err
	}
	environment := []string{"DB_CONNECTION=mysql", "DB_DSN=" + dsn}
	run := func(args ...string) (string, error) {
		output, runErr := invoke(ctx, executable, root, environment, args...)
		if runErr != nil {
			return "", fmt.Errorf("%s: %w", name, runErr)
		}
		return strings.TrimSpace(output), nil
	}
	if isMySQLMigrationOption(name) {
		return runMySQLMigrationOption(ctx, db, executable, root, dsn, name)
	}
	if name != "migrate" && name != "migrate-install" {
		if _, err := run("migrate"); err != nil {
			return Result{}, err
		}
	}
	command := map[string]string{
		"migrate": "migrate", "migrate-install": "migrate:install", "migrate-status": "migrate:status",
		"migrate-rollback": "migrate:rollback", "migrate-reset": "migrate:reset", "migrate-refresh": "migrate:refresh",
		"migrate-fresh": "migrate:fresh", "db-seed": "db:seed",
	}[name]
	if command == "" {
		return Result{}, fmt.Errorf("commands demo: unknown MySQL migration case %q", name)
	}
	if name == "migrate-fresh" {
		if _, err := db.ExecContext(ctx, "CREATE TABLE commands_demo_extra (id BIGINT PRIMARY KEY)"); err != nil {
			return Result{}, fmt.Errorf("prepare extra table for fresh: %w", err)
		}
	}
	args := []string{command}
	if name == "db-seed" {
		args = append(args, "--class=CommandsDemoSeeder")
	}
	output, err := run(args...)
	if err != nil {
		return Result{}, err
	}
	if err := verifyMigrationState(ctx, db, name, output); err != nil {
		return Result{}, err
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: output}, nil
}

func prepareMigrationSources(root string) error {
	path := filepath.Join(root, "database", "migrations")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("prepare migration source directory: %w", err)
	}
	seedPath := filepath.Join(root, "database", "seeders")
	if err := os.MkdirAll(seedPath, 0o755); err != nil {
		return fmt.Errorf("prepare seeder source directory: %w", err)
	}
	for _, name := range migrationSourceNames() {
		if err := os.WriteFile(filepath.Join(path, name), []byte("package migrations\n"), 0o644); err != nil {
			return fmt.Errorf("prepare registered migration %s: %w", name, err)
		}
	}
	return nil
}

func verifyMigrationState(ctx context.Context, db *sql.DB, name, output string) error {
	var count int
	if name != "migrate-install" {
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations").Scan(&count); err != nil {
			return fmt.Errorf("read %s migration count: %w", name, err)
		}
	}
	wantCount := len(migrationSourceNames())
	if name == "migrate-rollback" || name == "migrate-reset" {
		wantCount = 0
	}
	if name != "migrate-install" && count != wantCount {
		return fmt.Errorf("%s migration rows = %d, want %d; output = %q", name, count, wantCount, output)
	}
	tables := []string{"migrations", "users", "demo_orders", "demo_receipts"}
	if name == "migrate-rollback" || name == "migrate-reset" {
		tables = []string{"migrations"}
	}
	if name == "migrate-install" {
		tables = []string{"migrations"}
	}
	for _, table := range tables {
		exists, err := demoTableExists(ctx, db, table)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("%s table %q exists = false, want true; output = %q", name, table, output)
		}
	}
	if name == "migrate-rollback" || name == "migrate-reset" {
		for _, table := range []string{"users", "demo_orders", "demo_receipts"} {
			exists, err := demoTableExists(ctx, db, table)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("%s table %q exists = true, want false; output = %q", name, table, output)
			}
		}
	}
	if name == "migrate-status" && (!strings.Contains(output, "Ran?") || !strings.Contains(output, "202606050000_create_users_table")) {
		return fmt.Errorf("migrate-status output = %q, want applied migration table", output)
	}
	if name == "migrate-fresh" {
		exists, err := demoTableExists(ctx, db, "commands_demo_extra")
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("migrate-fresh extra table exists = true, want false")
		}
	}
	if name == "db-seed" {
		var marker string
		if err := db.QueryRowContext(ctx, "SELECT name FROM commands_demo_seeds WHERE id = 1").Scan(&marker); err != nil {
			return fmt.Errorf("read commands demo seed: %w", err)
		}
		if marker != "commands-demo" {
			return fmt.Errorf("db-seed marker = %q, want commands-demo", marker)
		}
	}
	return nil
}

func demoTableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", table).Scan(&count); err != nil {
		return false, fmt.Errorf("inspect commands demo table %s: %w", table, err)
	}
	return count == 1, nil
}

func clearDemoTables(ctx context.Context, db *sql.DB) (err error) {
	connection, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("open MySQL cleanup connection: %w", err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
		return fmt.Errorf("disable MySQL foreign key checks: %w", err)
	}
	defer func() {
		if _, restoreErr := connection.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=1"); restoreErr != nil {
			err = errors.Join(err, fmt.Errorf("restore MySQL foreign key checks: %w", restoreErr))
		}
	}()
	viewRows, err := connection.QueryContext(ctx, "SELECT table_name FROM information_schema.views WHERE table_schema = DATABASE()")
	if err != nil {
		return fmt.Errorf("list commands demo MySQL views: %w", err)
	}
	var views []string
	for viewRows.Next() {
		var view string
		if scanErr := viewRows.Scan(&view); scanErr != nil {
			viewRows.Close()
			return fmt.Errorf("scan commands demo MySQL view: %w", scanErr)
		}
		views = append(views, view)
	}
	if err := viewRows.Err(); err != nil {
		viewRows.Close()
		return fmt.Errorf("iterate commands demo MySQL views: %w", err)
	}
	if err := viewRows.Close(); err != nil {
		return fmt.Errorf("close commands demo MySQL view listing: %w", err)
	}
	for _, view := range views {
		quoted := "`" + strings.ReplaceAll(view, "`", "``") + "`"
		if _, err := connection.ExecContext(ctx, "DROP VIEW "+quoted); err != nil {
			return fmt.Errorf("drop commands demo MySQL view %s: %w", view, err)
		}
	}
	rows, err := connection.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return fmt.Errorf("list commands demo MySQL tables: %w", err)
	}
	var tables []string
	for rows.Next() {
		var table string
		if scanErr := rows.Scan(&table); scanErr != nil {
			rows.Close()
			return fmt.Errorf("scan commands demo MySQL table: %w", scanErr)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate commands demo MySQL tables: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close commands demo MySQL table listing: %w", err)
	}
	for _, table := range tables {
		quoted := "`" + strings.ReplaceAll(table, "`", "``") + "`"
		if _, err := connection.ExecContext(ctx, "DROP TABLE "+quoted); err != nil {
			return fmt.Errorf("drop commands demo MySQL table %s: %w", table, err)
		}
	}
	return nil
}
