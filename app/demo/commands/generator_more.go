package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runMoreGenerators(ctx context.Context, executable, root, name string) (Result, error) {
	switch name {
	case "listener-mode-conflict":
		return rejectGeneratorOptions(ctx, executable, root, name,
			[]string{"make:listener", "ConflictListener", "--queued", "--async"},
			"--async and --queued are mutually exclusive", "app/listeners/conflict_listener.go")
	case "stub-publish", "stub-override", "stub-force":
		return runStubScenario(ctx, executable, root, name)
	case "migration-registration":
		return runMigrationRegistration(ctx, executable, root)
	case "migration-realpath":
		path := filepath.Join(root, "custom", "migrations")
		return runGeneratorScenario(ctx, executable, root, name, generatorScenario{
			args:  []string{"make:migration", "create_audit_entries_table", "--create=audit_entries", "--path=" + path, "--realpath"},
			files: []generatedFile{{pattern: "custom/migrations/*_create_audit_entries_table.go", contains: []string{"Create(\"audit_entries\""}}},
		})
	case "migration-inference":
		return runMigrationInference(ctx, executable, root)
	}
	scenarios := map[string]generatorScenario{
		"listener-async":   {args: []string{"make:listener", "AsyncInvoiceListener", "--async"}, files: []generatedFile{{pattern: "app/listeners/async_invoice_listener.go", contains: []string{"Async() bool", "return true"}, excludes: []string{"ShouldQueue() bool"}}}},
		"listener-event":   {args: []string{"make:listener", "InvoicePaidListener", "--event=InvoicePaid"}, files: []generatedFile{{pattern: "app/listeners/invoice_paid_listener.go", contains: []string{"// Event hint: InvoicePaid"}}}},
		"migration-create": {args: []string{"make:migration", "create_invoices_table", "--create=invoices"}, files: []generatedFile{{pattern: "database/migrations/*_create_invoices_table.go", contains: []string{"schema.Bind(db).Create(\"invoices\"", "schema.Bind(db).DropIfExists(\"invoices\")"}}}},
		"migration-table":  {args: []string{"make:migration", "add_reference_to_invoices_table", "--table=invoices"}, files: []generatedFile{{pattern: "database/migrations/*_add_reference_to_invoices_table.go", contains: []string{"schema.Bind(db).Table(\"invoices\"", "DropColumn(\"reference\")"}}}},
		"migration-path":   {args: []string{"make:migration", "create_audit_entries_table", "--path=custom/migrations"}, files: []generatedFile{{pattern: "custom/migrations/*_create_audit_entries_table.go", contains: []string{"Create(\"audit_entries\""}}}},
	}
	scenario, ok := scenarios[name]
	if !ok {
		return Result{}, fmt.Errorf("commands demo: unknown generator case %q", name)
	}
	return runGeneratorScenario(ctx, executable, root, name, scenario)
}

func runGeneratorScenario(ctx context.Context, executable, root, name string, scenario generatorScenario) (Result, error) {
	output, err := invoke(ctx, executable, root, nil, scenario.args...)
	if err != nil {
		return Result{}, err
	}
	if !strings.Contains(output, "Created:") {
		return Result{}, fmt.Errorf("%s output = %q, want Created result", name, output)
	}
	for _, file := range scenario.files {
		if err := verifyGeneratedFile(root, file); err != nil {
			return Result{}, fmt.Errorf("%s: %w", name, err)
		}
	}
	return Result{Case: name, Command: strings.Join(scenario.args, " "), Output: strings.TrimSpace(output)}, nil
}

func rejectGeneratorOptions(ctx context.Context, executable, root, name string, args []string, want, absent string) (Result, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = root
	cmd.Env = commandEnvironment(nil)
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), want) {
		return Result{}, fmt.Errorf("%s exit error = %v, output = %q; want %q", name, err, output, want)
	}
	path := filepath.Join(root, filepath.FromSlash(absent))
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		return Result{}, fmt.Errorf("%s generated file stat = %v, want no file at %s", name, statErr, path)
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: strings.TrimSpace(string(output))}, nil
}

func runMigrationInference(ctx context.Context, executable, root string) (Result, error) {
	for _, scenario := range []generatorScenario{
		{args: []string{"make:migration", "create_payments_table"}, files: []generatedFile{{pattern: "database/migrations/*_create_payments_table.go", contains: []string{"Create(\"payments\""}}}},
		{args: []string{"make:migration", "add_reference_to_payments_table"}, files: []generatedFile{{pattern: "database/migrations/*_add_reference_to_payments_table.go", contains: []string{"Table(\"payments\"", "DropColumn(\"reference\")"}}}},
	} {
		if _, err := runGeneratorScenario(ctx, executable, root, "migration-inference", scenario); err != nil {
			return Result{}, err
		}
	}
	return Result{Case: "migration-inference", Command: "make:migration create_payments_table; make:migration add_reference_to_payments_table", Output: "verified inferred create and alter migration intents"}, nil
}

func runStubScenario(ctx context.Context, executable, root, name string) (Result, error) {
	args := []string{"stub:publish"}
	output, err := invoke(ctx, executable, root, nil, args...)
	if err != nil {
		return Result{}, err
	}
	stub := filepath.Join(root, "stubs", "event.stub")
	original, err := os.ReadFile(stub)
	if err != nil {
		return Result{}, fmt.Errorf("read published event stub: %w", err)
	}
	if !strings.Contains(output, "Created: stubs/event.stub") || !strings.Contains(string(original), "{{ .TypeName }}") {
		return Result{}, fmt.Errorf("stub:publish output = %q, content = %q; want published event template", output, original)
	}
	if name == "stub-publish" {
		return Result{Case: name, Command: strings.Join(args, " "), Output: "verified published event stub"}, nil
	}
	custom := []byte("package {{ .PackageName }}\n\ntype {{ .TypeName }} struct { CustomMarker bool }\n")
	if err := os.WriteFile(stub, custom, 0o644); err != nil {
		return Result{}, fmt.Errorf("write custom event stub: %w", err)
	}
	if name == "stub-override" {
		scenario := generatorScenario{args: []string{"make:event", "CustomEvent"}, files: []generatedFile{{pattern: "app/events/custom_event.go", contains: []string{"type CustomEvent struct{ CustomMarker bool }"}}}}
		return runGeneratorScenario(ctx, executable, root, name, scenario)
	}
	output, err = invoke(ctx, executable, root, nil, args...)
	if err != nil {
		return Result{}, err
	}
	retained, err := os.ReadFile(stub)
	if err != nil {
		return Result{}, fmt.Errorf("read skipped event stub: %w", err)
	}
	if !strings.Contains(output, "Skipped: stubs/event.stub") || string(retained) != string(custom) {
		return Result{}, fmt.Errorf("stub skip output = %q, content = %q; want retained custom stub", output, retained)
	}
	args = append(args, "--force")
	output, err = invoke(ctx, executable, root, nil, args...)
	if err != nil {
		return Result{}, err
	}
	restored, err := os.ReadFile(stub)
	if err != nil {
		return Result{}, fmt.Errorf("read overwritten event stub: %w", err)
	}
	if !strings.Contains(output, "Overwritten: stubs/event.stub") || string(restored) != string(original) {
		return Result{}, fmt.Errorf("stub force output = %q, content = %q; want restored built-in stub", output, restored)
	}
	return Result{Case: name, Command: "stub:publish; stub:publish; stub:publish --force", Output: "verified skip and force overwrite"}, nil
}

func runMigrationRegistration(ctx context.Context, executable, root string) (Result, error) {
	path := filepath.Join(root, "database", "migrations")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return Result{}, fmt.Errorf("prepare registered migration path: %w", err)
	}
	name := "202606050000_create_users_table.go"
	if err := os.WriteFile(filepath.Join(path, name), []byte("package migrations\n"), 0o644); err != nil {
		return Result{}, fmt.Errorf("prepare migration source: %w", err)
	}
	databasePath := filepath.Join(root, "registration.sqlite")
	output, err := invoke(ctx, executable, root, []string{"DB_DATABASE=" + databasePath}, "migrate:status")
	if err != nil {
		return Result{}, err
	}
	if !strings.Contains(output, "migration table not found") {
		return Result{}, fmt.Errorf("migration registration output = %q, want repository warning", output)
	}
	if _, err := invoke(ctx, executable, root, []string{"DB_DATABASE=" + databasePath}, "migrate:install"); err != nil {
		return Result{}, err
	}
	output, err = invoke(ctx, executable, root, []string{"DB_DATABASE=" + databasePath}, "migrate:status")
	if err != nil {
		return Result{}, err
	}
	if !strings.Contains(output, strings.TrimSuffix(name, ".go")) {
		return Result{}, fmt.Errorf("migration registration output = %q, want %s", output, name)
	}
	return Result{Case: "migration-registration", Command: "migrate:install; migrate:status", Output: strings.TrimSpace(output)}, nil
}
