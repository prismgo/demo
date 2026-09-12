package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type generatedFile struct {
	pattern  string
	contains []string
	excludes []string
}

type generatorScenario struct {
	args  []string
	files []generatedFile
}

func runGeneratorOptions(ctx context.Context, executable, root, name string) (Result, error) {
	if name == "model-unsupported-options" {
		return runUnsupportedModelOption(ctx, executable, root, name)
	}
	scenario, ok := generatorScenarios()[name]
	if !ok {
		return Result{}, fmt.Errorf("commands demo: unknown generator case %q", name)
	}
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
	if name == "model-seeder" {
		aliasArgs := []string{"make:model", "AliasInvoice", "--seed"}
		aliasOutput, err := invoke(ctx, executable, root, nil, aliasArgs...)
		if err != nil {
			return Result{}, fmt.Errorf("verify --seed alias: %w", err)
		}
		if err := verifyGeneratedFile(root, generatedFile{pattern: "database/seeders/alias_invoice_seeder.go", contains: []string{"func Seed(db *gorm.DB) error"}}); err != nil {
			return Result{}, fmt.Errorf("verify --seed alias: %w", err)
		}
		output += aliasOutput
		return Result{Case: name, Command: strings.Join(scenario.args, " ") + "; " + strings.Join(aliasArgs, " "), Output: strings.TrimSpace(output)}, nil
	}
	return Result{Case: name, Command: strings.Join(scenario.args, " "), Output: strings.TrimSpace(output)}, nil
}

func verifyGeneratedFile(root string, file generatedFile) error {
	matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(file.pattern)))
	if err != nil {
		return fmt.Errorf("match generated path %q: %w", file.pattern, err)
	}
	if len(matches) != 1 {
		return fmt.Errorf("generated path %q matches = %v, want exactly one file", file.pattern, matches)
	}
	content, err := os.ReadFile(matches[0])
	if err != nil {
		return fmt.Errorf("read generated file %s: %w", matches[0], err)
	}
	if len(content) == 0 {
		return fmt.Errorf("generated file %s is empty", matches[0])
	}
	for _, want := range file.contains {
		if !strings.Contains(string(content), want) {
			return fmt.Errorf("generated file %s missing %q; actual = %q", matches[0], want, content)
		}
	}
	for _, unwanted := range file.excludes {
		if strings.Contains(string(content), unwanted) {
			return fmt.Errorf("generated file %s contains %q, want absent; actual = %q", matches[0], unwanted, content)
		}
	}
	return nil
}

func runUnsupportedModelOption(ctx context.Context, executable, root, name string) (Result, error) {
	args := []string{"make:model", "UnsupportedInvoice", "--factory"}
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = root
	cmd.Env = commandEnvironment(nil)
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "unsupported option --factory for make:model") {
		return Result{}, fmt.Errorf("%s exit error = %v, output = %q; want unsupported --factory rejection", name, err, output)
	}
	path := filepath.Join(root, "app/models/unsupported_invoice.go")
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		return Result{}, fmt.Errorf("%s generated file stat = %v, want no file", name, statErr)
	}
	return Result{Case: name, Command: strings.Join(args, " "), Output: strings.TrimSpace(string(output))}, nil
}

func generatorScenarios() map[string]generatorScenario {
	return map[string]generatorScenario{
		"make-listener":       {args: []string{"make:listener", "InvoiceListener"}, files: []generatedFile{{pattern: "app/listeners/invoice_listener.go", contains: []string{"type InvoiceListener struct", "Handle(ctx context.Context, ev event.Event) error"}}}},
		"make-middleware":     {args: []string{"make:middleware", "EnsureAccount"}, files: []generatedFile{{pattern: "app/http/middleware/ensure_account.go", contains: []string{"func EnsureAccount() gin.HandlerFunc"}}}},
		"make-migration":      {args: []string{"make:migration", "create_invoices_table"}, files: []generatedFile{{pattern: "database/migrations/*_create_invoices_table.go", contains: []string{"Create(\"invoices\"", "DropIfExists(\"invoices\")"}}}},
		"make-model":          {args: []string{"make:model", "Invoice"}, files: []generatedFile{{pattern: "app/models/invoice.go", contains: []string{"type Invoice struct", "CreatedAt time.Time"}}}},
		"make-provider":       {args: []string{"make:provider", "BillingProvider"}, files: []generatedFile{{pattern: "app/providers/billing_provider.go", contains: []string{"type BillingProvider struct", "Register(app providercontract.Application) error"}}}},
		"make-resource":       {args: []string{"make:resource", "InvoiceResource"}, files: []generatedFile{{pattern: "app/http/resources/invoice_resource.go", contains: []string{"type InvoiceResource struct", "ToMap() map[string]any"}}}},
		"make-seeder":         {args: []string{"make:seeder", "InvoiceSeeder"}, files: []generatedFile{{pattern: "database/seeders/invoice_seeder.go", contains: []string{"database.RegisterSeeder(Seed)", "func Seed(db *gorm.DB) error"}}}},
		"model-migration":     {args: []string{"make:model", "Invoice", "--migration"}, files: []generatedFile{{pattern: "app/models/invoice.go", contains: []string{"type Invoice struct"}}, {pattern: "database/migrations/*_create_invoices_table.go", contains: []string{"Create(\"invoices\""}}}},
		"model-controller":    {args: []string{"make:model", "Invoice", "--controller"}, files: []generatedFile{{pattern: "app/models/invoice.go", contains: []string{"type Invoice struct"}}, {pattern: "app/http/controllers/invoice_controller.go", contains: []string{"type InvoiceController struct", "Index(ctx *gin.Context)"}}}},
		"model-resource":      {args: []string{"make:model", "Invoice", "--resource"}, files: []generatedFile{{pattern: "app/models/invoice.go", contains: []string{"type Invoice struct"}}, {pattern: "app/http/resources/invoice_resource.go", contains: []string{"ToMap() map[string]any"}}}},
		"model-seeder":        {args: []string{"make:model", "Invoice", "--seeder"}, files: []generatedFile{{pattern: "app/models/invoice.go", contains: []string{"type Invoice struct"}}, {pattern: "database/seeders/invoice_seeder.go", contains: []string{"database.RegisterSeeder(Seed)"}}}},
		"model-api":           {args: []string{"make:model", "Invoice", "--api"}, files: []generatedFile{{pattern: "app/models/invoice.go", contains: []string{"type Invoice struct"}}, {pattern: "app/http/controllers/invoice_controller.go", contains: []string{"Store(ctx *gin.Context)", "Destroy(ctx *gin.Context)"}, excludes: []string{"Create(ctx *gin.Context)", "Edit(ctx *gin.Context)"}}}},
		"model-table":         {args: []string{"make:model", "Invoice", "--table=billing_invoices"}, files: []generatedFile{{pattern: "app/models/invoice.go", contains: []string{"func (Invoice) TableName() string", "return \"billing_invoices\""}}}},
		"model-composition":   {args: []string{"make:model", "Invoice", "--migration", "--controller", "--resource", "--seeder", "--api"}, files: []generatedFile{{pattern: "app/models/invoice.go", contains: []string{"type Invoice struct"}}, {pattern: "database/migrations/*_create_invoices_table.go", contains: []string{"Create(\"invoices\""}}, {pattern: "app/http/controllers/invoice_controller.go", contains: []string{"Store(ctx *gin.Context)"}}, {pattern: "app/http/resources/invoice_resource.go", contains: []string{"ToMap() map[string]any"}}, {pattern: "database/seeders/invoice_seeder.go", contains: []string{"database.RegisterSeeder(Seed)"}}}},
		"controller-model":    {args: []string{"make:controller", "InvoiceController", "--model=Invoice"}, files: []generatedFile{{pattern: "app/http/controllers/invoice_controller.go", contains: []string{"TODO: connect Invoice manually"}, excludes: []string{"app/models", "prismgo-demo/app/models"}}}},
		"controller-api":      {args: []string{"make:controller", "InvoiceController", "--api"}, files: []generatedFile{{pattern: "app/http/controllers/invoice_controller.go", contains: []string{"Store(ctx *gin.Context)", "Destroy(ctx *gin.Context)"}, excludes: []string{"Create(ctx *gin.Context)", "Edit(ctx *gin.Context)"}}}},
		"controller-resource": {args: []string{"make:controller", "InvoiceController", "--resource"}, files: []generatedFile{{pattern: "app/http/controllers/invoice_controller.go", contains: []string{"Create(ctx *gin.Context)", "Edit(ctx *gin.Context)", "Destroy(ctx *gin.Context)"}}}},
		"command-signature":   {args: []string{"make:command", "SendInvoice", "--command=billing:send-invoice"}, files: []generatedFile{{pattern: "app/cmd/send_invoice.go", contains: []string{"console.MustDefinition(\"billing:send-invoice\""}}}},
		"listener-queued":     {args: []string{"make:listener", "InvoiceListener", "--queued"}, files: []generatedFile{{pattern: "app/listeners/invoice_listener.go", contains: []string{"ShouldQueue() bool", "return true"}, excludes: []string{"Async() bool"}}}},
	}
}
