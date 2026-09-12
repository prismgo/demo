package commands

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func isMySQLMigrationOption(name string) bool {
	switch name {
	case "database-option", "migration-force", "migration-path-option", "migration-realpath-option", "migration-pretend",
		"migration-seed", "migration-seeder", "migrate-step", "rollback-step", "rollback-batch", "fresh-drop-views", "seed-class":
		return true
	default:
		return false
	}
}

func runMySQLMigrationOption(ctx context.Context, db *sql.DB, executable, root, dsn, name string) (Result, error) {
	environment := []string{"DB_CONNECTION=mysql", "DB_DSN=" + dsn}
	if name == "database-option" {
		environment = append(environment, "DB_CONNECTION=sqlite", "DB_DATABASE="+filepath.Join(root, "alternate.sqlite"))
	}
	if name == "migration-force" {
		environment = append(environment, "APP_ENV=production")
	}
	run := func(args ...string) (string, error) {
		output, err := invoke(ctx, executable, root, environment, args...)
		if err != nil {
			return "", fmt.Errorf("%s: %w", name, err)
		}
		return strings.TrimSpace(output), nil
	}
	args := []string{"migrate"}
	switch name {
	case "database-option":
		args = append(args, "--database=mysql")
	case "migration-force":
		args = append(args, "--force")
	case "migration-path-option":
		first := filepath.Join(root, "database", "first")
		second := filepath.Join(root, "database", "second")
		if err := os.MkdirAll(first, 0o755); err != nil {
			return Result{}, fmt.Errorf("prepare first migration path: %w", err)
		}
		if err := os.MkdirAll(second, 0o755); err != nil {
			return Result{}, fmt.Errorf("prepare second migration path: %w", err)
		}
		names := migrationSourceNames()
		for index, source := range names {
			directory := first
			if index > 0 {
				directory = second
			}
			if err := os.Rename(filepath.Join(root, "database", "migrations", source), filepath.Join(directory, source)); err != nil {
				return Result{}, fmt.Errorf("move migration %s: %w", source, err)
			}
		}
		args = append(args, "--path=database/first", "--path=database/second")
	case "migration-realpath-option":
		args = append(args, "--path="+filepath.Join(root, "database", "migrations"), "--realpath")
	case "migration-pretend":
		args = append(args, "--pretend")
	case "migration-seed":
		args = append(args, "--seed", "--seeder=CommandsDemoSeeder")
	case "migration-seeder":
		args = append(args, "--seed", "--seeder=CommandsDemoSeeder")
	case "migrate-step":
		args = append(args, "--step")
	case "rollback-step", "rollback-batch":
		if _, err := run("migrate", "--step"); err != nil {
			return Result{}, err
		}
		args = []string{"migrate:rollback"}
		if name == "rollback-step" {
			args = append(args, "--step=1")
		} else {
			args = append(args, "--batch=2")
		}
	case "fresh-drop-views":
		if _, err := run("migrate"); err != nil {
			return Result{}, err
		}
		if _, err := db.ExecContext(ctx, "CREATE VIEW commands_demo_view AS SELECT id FROM users"); err != nil {
			return Result{}, fmt.Errorf("prepare commands demo view: %w", err)
		}
		args = []string{"migrate:fresh", "--drop-views"}
	case "seed-class":
		if _, err := run("migrate"); err != nil {
			return Result{}, err
		}
		args = []string{"db:seed", "--class=CommandsDemoSeeder"}
	}
	output, err := run(args...)
	if err != nil {
		return Result{}, err
	}
	count, err := migrationRowCount(ctx, db)
	if err != nil {
		return Result{}, err
	}
	want := len(migrationSourceNames())
	switch name {
	case "migration-pretend":
		want = 0
	case "rollback-step", "rollback-batch":
		want--
	}
	if count != want {
		return Result{}, fmt.Errorf("%s migration rows = %d, want %d; output = %q", name, count, want, output)
	}
	switch name {
	case "migration-pretend":
		if !strings.Contains(output, "Pretend:") {
			return Result{}, fmt.Errorf("%s output = %q, want Pretend operation", name, output)
		}
		exists, err := demoTableExists(ctx, db, "users")
		if err != nil {
			return Result{}, err
		}
		if exists {
			return Result{}, fmt.Errorf("%s users table exists = true, want false", name)
		}
	case "migration-seed", "migration-seeder", "seed-class":
		var marker string
		if err := db.QueryRowContext(ctx, "SELECT name FROM commands_demo_seeds WHERE id = 1").Scan(&marker); err != nil {
			return Result{}, fmt.Errorf("%s read seeded marker: %w", name, err)
		}
		if marker != "commands-demo" {
			return Result{}, fmt.Errorf("%s marker = %q, want commands-demo", name, marker)
		}
	case "migrate-step":
		var batches int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT batch) FROM migrations").Scan(&batches); err != nil {
			return Result{}, fmt.Errorf("%s read batches: %w", name, err)
		}
		if batches != want {
			return Result{}, fmt.Errorf("%s batches = %d, want %d", name, batches, want)
		}
	case "rollback-step", "rollback-batch":
		var latest int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations WHERE migration = '202609080010_create_demo_receipts_table'").Scan(&latest); err != nil {
			return Result{}, fmt.Errorf("%s read latest migration: %w", name, err)
		}
		if name == "rollback-step" && latest != 0 {
			return Result{}, fmt.Errorf("%s latest migration rows = %d, want 0", name, latest)
		}
		if name == "rollback-batch" {
			var second int
			if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations WHERE batch = 2").Scan(&second); err != nil {
				return Result{}, fmt.Errorf("%s read second batch: %w", name, err)
			}
			if second != 0 {
				return Result{}, fmt.Errorf("%s second batch rows = %d, want 0", name, second)
			}
		}
	case "fresh-drop-views":
		var views int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.views WHERE table_schema = DATABASE() AND table_name = 'commands_demo_view'").Scan(&views); err != nil {
			return Result{}, fmt.Errorf("%s inspect view: %w", name, err)
		}
		if views != 0 {
			return Result{}, fmt.Errorf("%s view count = %d, want 0", name, views)
		}
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: output}, nil
}

func migrationRowCount(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations").Scan(&count); err != nil {
		return 0, fmt.Errorf("read commands demo migration rows: %w", err)
	}
	return count, nil
}
