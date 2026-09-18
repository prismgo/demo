package schemademo

import (
	"fmt"

	dbschema "github.com/prismgo/framework/database/schema"
	"gorm.io/gorm"
)

// whenHasColumnScenario verifies WhenTableHasColumn runs its callback only when
// the target column already exists.
func whenHasColumnScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_when_has_column"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("code", 32)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	executed := false
	if err := builder.WhenTableHasColumn(table, "code", func() error {
		executed = true
		return builder.Table(table, func(blueprint *dbschema.Blueprint) {
			blueprint.String("added", 32)
		})
	}); err != nil {
		return "", err
	}
	absent := false
	if err := builder.WhenTableHasColumn(table, "missing", func() error {
		absent = true
		return nil
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("executed=%t added=%t absent_executed=%t", executed, builder.HasColumn(table, "added"), absent), nil
}

// whenMissingColumnScenario verifies WhenTableDoesntHaveColumn runs its callback
// only when the target column is missing.
func whenMissingColumnScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_when_missing_column"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("code", 32)
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	present := false
	if err := builder.WhenTableDoesntHaveColumn(table, "code", func() error {
		present = true
		return nil
	}); err != nil {
		return "", err
	}
	missing := false
	if err := builder.WhenTableDoesntHaveColumn(table, "added", func() error {
		missing = true
		return builder.Table(table, func(blueprint *dbschema.Blueprint) {
			blueprint.String("added", 32)
		})
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("present_executed=%t missing_executed=%t added=%t", present, missing, builder.HasColumn(table, "added")), nil
}

// whenMissingIndexScenario verifies WhenTableDoesntHaveIndex runs its callback
// only when the target index is missing.
func whenMissingIndexScenario(db *gorm.DB) (string, error) {
	const table = "schema_demo_when_missing_index"
	const present = "idx_schema_demo_when_missing_index_status"
	const absent = "idx_schema_demo_when_missing_index_level"
	builder := dbschema.New(db)
	if err := builder.Create(table, func(blueprint *dbschema.Blueprint) {
		blueprint.Id()
		blueprint.String("status", 16)
		blueprint.String("level", 16)
		blueprint.IndexNamed(present, "status")
	}); err != nil {
		return "", err
	}
	defer func() { _ = builder.Drop(table) }()

	existing := false
	if err := builder.WhenTableDoesntHaveIndex(table, present, func() error {
		existing = true
		return nil
	}); err != nil {
		return "", err
	}
	missing := false
	if err := builder.WhenTableDoesntHaveIndex(table, []string{"level"}, func() error {
		missing = true
		return builder.Table(table, func(blueprint *dbschema.Blueprint) {
			blueprint.IndexNamed(absent, "level")
		})
	}); err != nil {
		return "", err
	}
	return fmt.Sprintf("existing_executed=%t missing_executed=%t created=%t", existing, missing, builder.HasIndex(table, absent)), nil
}
